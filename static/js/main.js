/* const playgroundsSection = document.querySelector('#playgrounds')
		// window.addEventListener('load', () => {
		// 	registerSW();
		// })
		fetch("/api/playgrounds")
			.then(res => res.json())
			.then(playgrounds => {
				// playgrounds.map(playground => {
				// 	const div = document.createElement('div')
				// 	div.textContent = playground.Name
				// 	playgroundsSection.appendChild(div)
				// })
				console.log(playgrounds)
			}); */

		// async function registerSW(params) {
		// 	if ('serviceWorker' in navigator) {
		// 		try {
		// 			await navigator.serviceWorker.register('sw.js');
		// 		} catch (e) {
		// 			console.log('SW registration failed');
		// 		}
		// 	}
		// }