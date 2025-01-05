# dtone-micro-sms

### Responsible
This project act as a sms management.

### Configuration
Please update related configuration in env/local-dev.env before running the program.
````
go run main.go
````

### Remark
1. Postgres Database connection need to be established.
2. Redis Cache connection need to be established.
3. Kafka client connection need to be established.
4. GRPC server service connection need to be started.

### Future Enhancement
1. Testing for the code and logic need to be done.
2. Stress test need to do.
3. Benchmarking need to do.
4. Hard code configuration need to be update dynamic.
5. Auto reload configuration action for every env changes.
6. Specific config file when program start up dynamically.
7. Auto recovery after panic.