# FireFly Simulator

FireFly Simulator is an interactive web-based simulation of firefly synchronization behavior. It demonstrates how simple rules can lead to complex, emergent patterns in nature.

## Features

- Real-time simulation of firefly flashing patterns
- WebSocket-based communication for live updates
- Customizable grid size and firefly behavior
- Responsive design that adapts to different screen sizes
- Manual and automatic simulation reset functionality
- Timer display showing time since last reset

## Technical Stack

- Backend: Go (Golang)
- Frontend: HTML, CSS, JavaScript
- Communication: WebSockets

## Getting Started

### Prerequisites

- Go 1.15 or higher
- A modern web browser

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/firefly-simulator.git
   cd firefly-simulator
   ```

2. Install the required Go packages:
   ```
   go get github.com/gorilla/websocket
   ```

3. Build and run the server:
   ```
   go build
   ./firefly-simulator
   ```

4. Open `index.html` in your web browser or serve it using a local HTTP server.

## Usage

- The simulation starts automatically when you open the web page.
- Click the "Restart Simulation" button to manually reset the simulation.
- The simulation automatically resets every hour.
- Observe how the fireflies start to synchronize their flashing over time.

## Customization

You can customize various aspects of the simulation by modifying the `main.go` file:

- Adjust `gridSize` to change the number of fireflies.
- Modify the `initializeState` function to change the initial distribution of fireflies.
- Alter the `cycleRate` and `flashDuration` in the `initializeState` function to change firefly behavior.
- Adjust the neighbor checking radius in the `checkNeighbors` function.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by the synchronization behavior of real fireflies
- Thanks to the Go community for the excellent WebSocket library