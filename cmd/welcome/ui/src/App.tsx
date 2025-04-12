import { useState } from "react";
import reactLogo from "./assets/icon.png";
import reactLogo2 from "./assets/icon_black.svg";
import "./App.css";

function Step({ children, buttonText, onButtonClick }: { children: React.ReactNode; buttonText: string; onButtonClick: () => void }) {
	return (
		<main className="bg-base-200 min-h-screen w-full flex items-center justify-center">
			<div className="p-6 flex min-h-screen flex-col items-center justify-between space-y-3">
				<div className="flex-none text-center flex flex-col items-center">{children}</div>
				<button
					type="button"
					className="btn btn-primary w-full flex-none"
					onClick={onButtonClick}
				>
					{buttonText}
				</button>
			</div>
		</main>
	);
}

function App() {
	const [currentStep, setCurrentStep] = useState(0);
	const [autoLaunch, setAutoLaunch] = useState(true);

	if (currentStep === 0) {
		return (
			<Step
				buttonText="Get Started"
				onButtonClick={() => setCurrentStep(1)}
			>
				<div className="flex flex-col items-center space-y-3">
					<img
						src={reactLogo}
						alt="Pareto Security Logo"
						className="h-52 w-52"
					/>
					<div className="text-center">
						<h1 className="text-3xl">Welcome to</h1>
						<h2 className="text-primary font-extrabold text-4xl">
							Pareto Security
						</h2>
					</div>
					<p className="text-sm text-justify text-content">
					Pareto Security is an app that regularly checks your Mac's security
					configuration. It helps you take care of 20% of security tasks that
					prevent 80% of problems.
				</p>
				<fieldset className="fieldset">
					<label className="fieldset-label">
						<input
							type="checkbox"
							checked={autoLaunch}
							onChange={() => setAutoLaunch(!autoLaunch)}
							className="checkbox checkbox-xs checkbox-primary"
						/>
						Automatically launch on system startup
					</label>
				</fieldset>
				</div>

			</Step>
		);
	}

	if (currentStep === 1) {
		return (
			<Step
				buttonText="Continue"
				onButtonClick={() => setCurrentStep(1)}
			>
				<div className="flex flex-col items-center space-y-3">
					<img
						src={reactLogo}
						alt="Pareto Security Logo"
						className="h-52 w-52"
					/>
					<h1 className="text-3xl">Done!</h1>
					<p className="text-sm text-justify text-content grow">
					Pareto Security is now running in the background. You can find the app by looking for{" "}
					<img
						src={reactLogo2}
						alt="Pareto Security Logo"
						className="h-6 w-6 inline-block"
					/>{" "}
					in the tray.
				</p>
				</div>

			</Step>
		);
	}
}

export default App;
