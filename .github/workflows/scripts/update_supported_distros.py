import requests
import os
import yaml
import re
import time
from typing import List, Dict, Any, Optional, Tuple
from ruamel.yaml import YAML
from ruamel.yaml.scalarstring import LiteralScalarString, DoubleQuotedScalarString
from ruamel.yaml.comments import CommentedSeq

"""
This script checks the currently supported versions of ubuntu, debian and fedora on https://endoflife.date/ 
and updates the .github/workflows/distro.yml and .github/workflows/build.yml files accordingly.

It automatically runs every Monday at 8:30 AM CEST, via a Github Action, which creates a pull request if any changes have been made.
It can also be run manually.
"""

IGNORED_DISTROS: List[str] = ['arch']

MAX_RETRIES = 3
RETRY_DELAY = 10

def get_supported_distros() -> Tuple[List[str], List[str]]:
    """
    Fetch currently supported distribution versions from endoflife.date API.
    Retries failed API calls up to MAX_RETRIES times with RETRY_DELAY seconds between attempts.

    Returns:
        Tuple containing:
        - List[str]: List of supported distros in format "distro-version" (e.g., "ubuntu-24.04")
        - List[str]: List of distro families where the API call failed after all retries
    """
    distros = ['debian', 'ubuntu', 'fedora']
    supported = []
    failed = []

    for distro in distros:
        last_error = None

        for attempt in range(1, MAX_RETRIES + 1):
            try:
                response = requests.get(f'https://endoflife.date/api/v1/products/{distro}', timeout=10)
                response.raise_for_status()
                releases = response.json()['result']['releases']

                for release in releases:
                    if release['isEol'] is False:
                        supported.append(f"{distro}-{release['name']}")

                last_error = None
                break

            except (requests.exceptions.RequestException, KeyError, TypeError) as e:
                last_error = e
                if attempt < MAX_RETRIES:
                    print(f"Error fetching data for {distro} (attempt {attempt}/{MAX_RETRIES}): {e}. Retrying in {RETRY_DELAY}s...")
                    time.sleep(RETRY_DELAY)

        if last_error is not None:
            print(f"Error fetching data for {distro}: failed after {MAX_RETRIES} attempts: {last_error}")
            failed.append(distro)

    return supported, failed


def get_current_distros() -> Optional[List[Dict[str, Any]]]:
    """Read the current distribution configurations from distro.yml file."""
    try:
        current = os.path.dirname(os.path.abspath(__file__))
        path = os.path.join(current, "..", "distro.yml")
        
        with open(path, 'r') as f:
            data = yaml.safe_load(f)
        
        matrix = data['jobs']['distro-tests']['strategy']['matrix']['include']
        distros = [d for d in matrix if d.get('distro') not in IGNORED_DISTROS]

        return distros
    
    except Exception as e:
        print(f"Error reading distro.yml file: {e}")
        return None


def fix_cleanup_run_field(content: str) -> str:
    """
    Fix YAML formatting for the Cleanup step's run field.
    
    The YAML formatter adds a '|-' in the run field in the Cleanup step,
    so this function replaces it with a '|' to preserve the original format.

    """
    pattern = r'(- name: Cleanup\s*\n\s*if: always\(\)\s*\n\s*run:\s*)(\|-)'
    return re.sub(pattern, r'\1|', content)


def extract_distro_template(distro_name: str) -> Optional[Dict[str, str]]:
    """
    Extract a template configuration for a given distribution type.
    
    Finds the first entry for the specified distro type and creates a template
    by replacing the version number with a placeholder.
    """
    try:
        current = os.path.dirname(os.path.abspath(__file__))
        path = os.path.join(current, "..", "distro.yml")
        
        with open(path, 'r') as f:
            data = yaml.safe_load(f)
        
        matrix = data['jobs']['distro-tests']['strategy']['matrix']['include']
        
        # Find the first entry for this distro type
        for entry in matrix:
            if entry.get('distro', '').startswith(distro_name + '-'):
                # Extract version number from the distro name
                version_pattern = entry['distro'].replace(distro_name + '-', '')
                
                # Create template by replacing the specific version with {version}
                template = {
                    'image': entry['image'].replace(version_pattern, '{version}'),
                    'setup': entry['setup'],
                    'installer': entry['installer'],
                    'verify_package': entry['verify_package']
                }
                
                return template
        
        # If no existing entry found, return None
        return None
        
    except Exception as e:
        print(f"Error extracting template for {distro_name}: {e}")
        return None


def remove_distro(distro: str) -> None:
    """Remove a distribution entry from the distro.yml matrix."""
    current = os.path.dirname(os.path.abspath(__file__))
    path = os.path.join(current, "..", "distro.yml")
    
    # Use ruamel.yaml to preserve formatting
    yaml = YAML()
    yaml.preserve_quotes = True
    yaml.default_flow_style = False
    yaml.width = 4096
    yaml.indent(mapping=2, sequence=4, offset=2)
    
    with open(path, 'r') as f:
        data = yaml.load(f)
    
    matrix = data['jobs']['distro-tests']['strategy']['matrix']['include']

    # Remove the distro entry
    for i, entry in enumerate(matrix):
        if entry.get('distro') == distro:
            matrix.pop(i)
            break

    # Write back preserving format
    with open(path, 'w') as f:
        yaml.dump(data, f)
    
    # Post-process to fix yaml formatting (Replaces |- with | for run fields)
    with open(path, 'r') as f:
        content = f.read()
    
    content = fix_cleanup_run_field(content)
    
    with open(path, 'w') as f:
        f.write(content)
        
    
def add_distro(distro: str) -> None:
    """
    Add a new distribution entry to the distro.yml matrix.
    
    Extracts a template from existing entries of the same distro type and
    inserts the new entry in the appropriate position to maintain ordering.
    """
    current = os.path.dirname(os.path.abspath(__file__))
    path = os.path.join(current, "..", "distro.yml")
    
    parts = distro.split('-')
    distro_name = parts[0]
    distro_version = '-'.join(parts[1:])
    
    template = extract_distro_template(distro_name)
    
    if not template:
        print(f"Warning: No existing template found for {distro_name}. Skipping.")
        return

    yaml = YAML()
    yaml.preserve_quotes = True
    yaml.default_flow_style = False
    yaml.width = 4096
    yaml.indent(mapping=2, sequence=4, offset=2)
    
    with open(path, 'r') as f:
        data = yaml.load(f)
        
    matrix = data['jobs']['distro-tests']['strategy']['matrix']['include']
    
    matrix_dict = {i: entry['distro'] for i, entry in enumerate(matrix)}
    
    # Find where to insert the new distro
    insert_position = len(matrix_dict)
    for idx in reversed(range(len(matrix_dict))):
        if matrix_dict[idx].startswith(distro_name + '-'):
            insert_position = idx + 1
            break
    
    # Update dictionary keys to make room for the new entry
    new_matrix_dict = {}
    for idx in range(len(matrix_dict)):
        if idx < insert_position:
            new_matrix_dict[idx] = matrix_dict[idx]
        else:
            new_matrix_dict[idx + 1] = matrix_dict[idx]
    
    # Add the new distro at the correct position
    new_matrix_dict[insert_position] = distro
    
    # Rebuild the matrix list
    new_matrix = []
    for idx in sorted(new_matrix_dict.keys()):
        if new_matrix_dict[idx] == distro:
            # Create new entry from template
            new_entry = {}
            new_entry['distro'] = distro
            new_entry['image'] = template['image'].format(version=distro_version)
            
            # Handle setup field for multiline
            setup_content = template['setup']
            
            # Convert newlines to actual newlines
            if '\\n' in setup_content:
                setup_content = setup_content.replace('\\n', '\n')
            new_entry['setup'] = LiteralScalarString(setup_content)
            
            # Preserve quotes for installer and verify_package
            new_entry['installer'] = DoubleQuotedScalarString(template['installer'])
            new_entry['verify_package'] = DoubleQuotedScalarString(template['verify_package'])
            
            new_matrix.append(new_entry)

        else:
            for entry in matrix:
                if entry['distro'] == new_matrix_dict[idx]:
                    new_matrix.append(entry)
                    break
    
    # Replace the matrix with the new one
    data['jobs']['distro-tests']['strategy']['matrix']['include'] = new_matrix

    with open(path, 'w') as f:
        yaml.dump(data, f)


def update_build_yml(supported_distros: List[str]) -> None:
    """ Update the distribution matrix in the build.yml file. """

    current = os.path.dirname(os.path.abspath(__file__))
    path = os.path.join(current, "..", "build.yml")
    
    new_matrix = []
    for d in supported_distros:
        new_matrix.append(d.replace('-', '/'))
    
    new_matrix.append('archlinux')
    
    yaml = YAML()
    yaml.preserve_quotes = True
    yaml.default_flow_style = None
    yaml.width = 4096
    yaml.indent(mapping=2, sequence=4, offset=2)
    
    with open(path, 'r') as f:
        data = yaml.load(f)
        
    inline_matrix = CommentedSeq(new_matrix)
    inline_matrix.fa.set_flow_style()
    
    data['jobs']['checks-on-linux']['strategy']['matrix']['distribution'] = inline_matrix
    
    with open(path, 'w') as f:
        yaml.dump(data, f)
    
    print(f"Updated build.yml matrix.")


def preserve_failed_families(supported_distros: List[str], failed_families: List[str], current_distro_names: List[str]) -> None:
    """Preserve current distro entries for families where the API call failed."""
    print(f"Warning: API calls failed for: {', '.join(failed_families)}. Preserving current entries.")
    for distro in current_distro_names:
        family = distro.split('-')[0]
        if family in failed_families and distro not in supported_distros:
            supported_distros.append(distro)


def set_github_output(name: str, value: str) -> None:
    """
    Set output variables for GitHub Actions.
    
    Writes output variables to the GITHUB_OUTPUT file in the format expected
    by GitHub Actions. Handles multi-line values using delimiters.
    """
    
    github_output = os.environ.get('GITHUB_OUTPUT')
    if github_output:
        with open(github_output, 'a') as f:

            if '\n' in value:
                import uuid
                delimiter = f"EOF_{uuid.uuid4().hex}"
                f.write(f"{name}<<{delimiter}\n{value}\n{delimiter}\n")
            else:
                f.write(f"{name}={value}\n")


def main() -> None:
    """
    Main entry point for the distro update script.
    
    Compares currently supported distros from endoflife.date with the
    distro.yml matrix and updates it as needed. Also updates the build.yml
    matrix. Creates GitHub Actions outputs for the workflow to use.
    """

    supported_distros, failed_families = get_supported_distros()
    current_distros = get_current_distros()
    current_distro_names = [d['distro'] for d in current_distros]

    if failed_families:
        preserve_failed_families(supported_distros, failed_families, current_distro_names)

    if set(supported_distros) != set(current_distro_names):
        print("Updating supported distros in .github/workflows/distro.yml")
        
        added_distros = []
        removed_distros = []
        
        for distro in current_distro_names:
            if distro not in supported_distros:
                print(f"The {distro} distro is not supported, removing from the matrix.")
                removed_distros.append(distro)
                remove_distro(distro)
        
        for distro in supported_distros:
            if distro not in current_distro_names:
                print(f"Adding new distro version: {distro} to the matrix.")
                added_distros.append(distro)
                add_distro(distro)

        print("Updating supported distros in .github/workflows/build.yml")
        update_build_yml(supported_distros)

        set_github_output("changes_made", "true")
        
        pr_description = "Update supported Linux distribution versions based on https://endoflife.date/\n\n"
        
        if added_distros:
            pr_description += "### Added distros:\n"
            for distro in added_distros:
                pr_description += f"- {distro}\n"
            pr_description += "\n"
        
        if removed_distros:
            pr_description += "### Removed distros (reached end-of-life):\n"
            for distro in removed_distros:
                pr_description += f"- {distro}\n"
            pr_description += "\n"
        
        set_github_output("pr_description", pr_description.rstrip())

    else:
        print("Supported distros are up to date.")
        set_github_output("changes_made", "false")


if __name__ == "__main__":
    main()