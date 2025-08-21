import requests
import os
import yaml
from ruamel.yaml import YAML
from ruamel.yaml.scalarstring import LiteralScalarString, DoubleQuotedScalarString
import re

"""
This script checks the currently supported versions of ubuntu, debian and fedora on https://endoflife.date/ 
and updates the .github/workflows/distro.yml file accordingly.

It automatically runs every Monday at 8:30 AM CEST, via a Github Action, which creates a pull request if any changes have been made.
It can also be run manually.
"""

IGNORED_DISTROS = ['arch']

DISTRO_TEMPLATES = {
    'ubuntu': {
        'image': 'jrei/systemd-ubuntu:{version}',
        'setup': '''apt-get update
apt-get upgrade -y''',
        'installer': 'apt-get install -y',
        'verify_package': 'dpkg -l paretosecurity'
    },
    'debian': {
        'image': 'jrei/systemd-debian:{version}',
        'setup': '''apt-get update
apt-get upgrade -y''',
        'installer': 'apt-get install -y',
        'verify_package': 'dpkg -l paretosecurity'
    },
    'fedora': {
        'image': 'jrei/systemd-fedora:{version}',
        'setup': '''dnf -y update
dnf -y upgrade
dnf -y install which''',
        'installer': 'dnf -y install',
        'verify_package': 'rpm -q paretosecurity'
    }
}

def get_supported_distros():

    distros = ['debian', 'ubuntu', 'fedora']
    supported = []

    for distro in distros:

        response = requests.get(f'https://endoflife.date/api/v1/products/{distro}')
        releases = response.json()['result']['releases']

        for release in releases:
            if release['isEol'] is False:
                supported.append(f"{distro}-{release['name']}")

    return supported


def get_current_distros():
    try:
        current = os.path.dirname(os.path.abspath(__file__))
        path = os.path.join(current, "..", ".github", "workflows", "distro.yml")
        
        with open(path, 'r') as f:
            data = yaml.safe_load(f)
        
        matrix = data['jobs']['distro-tests']['strategy']['matrix']['include']
        distros = [d for d in matrix if d.get('distro') not in IGNORED_DISTROS]

        return distros
    
    except Exception as e:
        print(f"Error reading distro.yml file: {e}")
        return None


def remove_distro(distro):
    current = os.path.dirname(os.path.abspath(__file__))
    path = os.path.join(current, "..", ".github", "workflows", "distro.yml")
    
    yaml_parser = YAML()
    yaml_parser.preserve_quotes = True
    yaml_parser.width = 4096  
    yaml_parser.indent(mapping=2, sequence=4, offset=2)
    
    with open(path, 'r') as f:
        data = yaml_parser.load(f)
    
    matrix = data['jobs']['distro-tests']['strategy']['matrix']['include']
    
    for i, entry in enumerate(matrix):
        if entry.get('distro') == distro:
            matrix.pop(i)
            break
    
    # Write the new matrix back to the distro.yml file
    with open(path, 'w') as f:
        yaml_parser.dump(data, f)
    
    # Post-process to fix yaml_parser formatting (Replaces |- with | for run fields)
    with open(path, 'r') as f:
        content = f.read()

    content = re.sub(r'(\s+run:\s*)(\|-)(\s*\n)', r'\1|\3', content)
    
    with open(path, 'w') as f:
        f.write(content)
        
    
def add_distro(distro):
    current = os.path.dirname(os.path.abspath(__file__))
    path = os.path.join(current, "..", ".github", "workflows", "distro.yml")
    
    parts = distro.split('-')
    distro_name = parts[0]
    distro_version = '-'.join(parts[1:])
    
    template = DISTRO_TEMPLATES[distro_name]
    
    yaml_parser = YAML()
    yaml_parser.preserve_quotes = True
    yaml_parser.width = 4096  
    yaml_parser.indent(mapping=2, sequence=4, offset=2)
    
    with open(path, 'r') as f:
        data = yaml_parser.load(f)
    
    matrix = data['jobs']['distro-tests']['strategy']['matrix']['include']
    
    # Find where to insert the new distro (after the last occurrence of the same distro type). If none found, append to the end.
    insert_index = len(matrix)
    for i, entry in enumerate(matrix):
        if entry.get('distro', '').startswith(distro_name + '-'):
            # Find the last occurrence of this distro type
            for j in range(i + 1, len(matrix)):
                if not matrix[j].get('distro', '').startswith(distro_name + '-'):
                    insert_index = j
                    break
            else:
                insert_index = len(matrix)
    
    # Create new entry from template
    new_entry = {
        'distro': distro,
        'image': template['image'].format(version=distro_version),
        'setup': template['setup'],
        'installer': template['installer'],
        'verify_package': template['verify_package']
    }
    
    # Convert multi-line strings to proper YAML format
    new_entry['setup'] = LiteralScalarString(new_entry['setup'] + '\n')
    # Preserve double quotes
    new_entry['installer'] = DoubleQuotedScalarString(new_entry['installer'])
    new_entry['verify_package'] = DoubleQuotedScalarString(new_entry['verify_package'])
    
    # Insert the new entry at the calculated position
    matrix.insert(insert_index, new_entry)
        
    # Write the file back to the distro.yml file
    with open(path, 'w') as f:
        yaml_parser.dump(data, f)
    
    # Post-process to fix yaml_parser formatting (Replaces |- with | for run fields)
    with open(path, 'r') as f:
        content = f.read()
    
    content = re.sub(r'(\s+run:\s*)(\|-)(\s*\n)', r'\1|\3', content)
    
    with open(path, 'w') as f:
        f.write(content)
    

def main():

    supported_distros = get_supported_distros()
    current_distros = get_current_distros()
    current_distro_names = [d['distro'] for d in current_distros]

    if set(supported_distros) != set(current_distro_names):
        print("Updating supported distros in .github/workflows/distro.yml")
        
        for distro in current_distro_names:
            if distro not in supported_distros:
                print(f"The {distro} distro is not supported, removing from the matrix.")
                remove_distro(distro)
        
        for distro in supported_distros:
            if distro not in current_distro_names:
                print(f"Adding new distro version: {distro} to the matrix.")
                add_distro(distro)

    else:
        print("Supported distros are up to date.")


if __name__ == "__main__":
    main()
