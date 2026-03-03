# win-powerctl

The laziest way to shut down your PCThe laziest way to shut down your PC

## Usage

1. Clone the repository:

   ```bash
   git clone https://github.com/aliffatulmf/win-powerctl.git
   ```

2. Navigate to the project directory:

   ```bash
   cd win-powerctl
   ```

3. Build the project:

   ```bash
   go build ./cmd/win-powerctl
   ```

4. Run the executable:

   ```bash
   ./win-powerctl
   ```

5. Flags:
   - `install`

     Sets up the application as a Windows service.

     ```bash
     ./win-powerctl install
     ```

   - `uninstall`

     Removes the application from the system as a service.

     ```bash
     ./win-powerctl uninstall
     ```

   - `service`

     Runs the application in service mode.

     ```bash
     ./win-powerctl service
     ```

6. API Endpoints:
   - `GET /shutdown`

     Triggers a graceful system shutdown.

     ```bash
     curl http://localhost:10125/shutdown
     ```

   - `GET /health`

     Checks the application's health status.

     ```bash
     curl http://localhost:10125/health
     ```
