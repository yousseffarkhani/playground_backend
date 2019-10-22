let deferredPrompt; // Allows to show the install prompt
let setupButton1;
let setupButton2;
let setupDiv;

window.addEventListener('beforeinstallprompt', (e) => {
	// Prevent Chrome 67 and earlier from automatically showing the prompt
	e.preventDefault();
	// Stash the event so it can be triggered later.
	deferredPrompt = e;
	console.log("beforeinstallprompt fired");
	displayButton(setupButton1, "setup_button1");
	displayButton(setupButton2, "setup_button2");
	displayButton(setupDiv, "setup_div");
});

function displayButton(button, ID){
	if (button == undefined) {
		button = document.getElementById(ID);
	}
	// Show the setup button
	button.style.display = "block";
	button.disabled = false;
}

function installApp() {
    // Show the prompt
    deferredPrompt.prompt();
    setupButton1.disabled = true;
    setupButton2.disabled = true;
    setupDiv.disabled = true;
    // Wait for the user to respond to the prompt
    deferredPrompt.userChoice
        .then((choiceResult) => {
            if (choiceResult.outcome === 'accepted') {
                console.log('PWA setup accepted');
                // hide our user interface that shows our A2HS button
                setupButton1.style.display = 'none';
				setupButton2.style.display = 'none';
				setupDiv.style.display = 'none';
            } else {
                console.log('PWA setup rejected');
            }
            deferredPrompt = null;
        });
}

window.addEventListener('appinstalled', (evt) => {
    console.log("appinstalled fired", evt);
});

window.addEventListener('load', () => {
	registerSW();
})
async function registerSW() {
	if ('serviceWorker' in navigator) {
		try {
			await navigator.serviceWorker.register('/sw.js');
		} catch (e) {
			console.log('SW registration failed');
		}
	}
}
