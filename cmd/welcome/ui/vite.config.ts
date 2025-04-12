import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import elmPlugin from "vite-plugin-elm";
import { run as runEOL2 } from "elm-optimize-level-2";
import fs from "node:fs";
import { resolve } from "node:path";
import { temporaryFile } from "tempy";

const compileWithEOL2 = async (targets: string[]) => {
	const output = temporaryFile({ extension: "elm" });
	await runEOL2({
		inputFilePath: targets,
		outputFilePath: output,
		optimizeSpeed: true,
		processOpts: { stdio: ["inherit", "ignore", "inherit"] },
	});
	return fs.readFileSync(output).toString();
};

// https://vitejs.dev/config/
export default defineConfig(async () => ({
	plugins: [
		elmPlugin({
			optimize: false,
      debug: true,
		}),
		tailwindcss(),
	],
	build: {
		rollupOptions: {
			input: {
				main: resolve(__dirname, "index.html"),
			},
		},
	},
}));
