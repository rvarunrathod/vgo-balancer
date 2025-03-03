# vgo-balancer

vgo-balancer is a Go-based load balancing solution designed to distribute network or application traffic across multiple servers efficiently. By using vgo-balancer, you can enhance the availability and reliability of your services, ensuring optimal performance and fault tolerance.

## Features

- **Efficient Load Distribution**: Balances incoming traffic across multiple servers to prevent overload on a single server.
- **High Availability**: Ensures continuous service availability by redirecting traffic from failed or overloaded servers to healthy ones.

## Installation

To install vgo-balancer, ensure you have [Go](https://golang.org/dl/) installed on your system. Then, run:

```bash
go get -u github.com/rvarunrathod/vgo-balancer
```

## Usage

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/rvarunrathod/vgo-balancer.git
   cd vgo-balancer
   ```

2. **Configuration**:

   - Modify the `config.yaml` file to define your backend servers and load balancing preferences.

3. **Build and Run**:

   - To build the application:

     ```bash
     go build -o vgo-balancer cmd/main.go
     ```

   - To run the application:

     ```bash
     ./vgo-balancer
     ```

## Configuration

The `config.yaml` file allows you to specify:

- **Backend Servers**: List of servers to distribute traffic to.
- **Load Balancing Algorithm**: Choose from available algorithms like round-robin, weighted-round-robin, ip-hash, least-response-time.
- **Health Check Parameters**: Define health check intervals and failure thresholds to monitor server health.

### Docker Compose

If you have docker running in your local you can test there.

This command will build docker images for you,
```bash
docker compose build
```

And Run below command to run load balancer and servers 
```bash
docker compose up
```

## Benchmarking

A benchmarking script is included in the `tools/benchmark` directory. Run it with:

```bash
cd tools/benchmark && go run main.go -url http://localhost:8080 -c 10 -n 1000 -d 10s
```

Available flags:
- `-url`: Target URL (default: "http://localhost:8080")
- `-c`: Number of concurrent requests (default: 10)
- `-n`: Total number of requests (default: 1000)
- `-d`: Duration of the test (e.g., "30s", "5m")

## Contributing

We welcome contributions! Please fork the repository and submit a pull request with your enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
