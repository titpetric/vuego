class VuegoApp {
  constructor(containerId = "container") {
    this.container = document.getElementById(containerId)
    if (this.container) {
      this.container.innerHTML = "Filled from JavaScript"
    }
  }
}

document.addEventListener('DOMContentLoaded', () => new VuegoApp());
