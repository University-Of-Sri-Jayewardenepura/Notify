# Build stage
FROM ballerina/ballerina:2201.10.0 AS build

WORKDIR /app

# Copy source files
COPY Ballerina.toml .
COPY main.bal .

# Build the application
RUN bal build

# Runtime stage
FROM ballerina/ballerina:2201.10.0-runtime

WORKDIR /home/ballerina

# Copy the built JAR from build stage
COPY --from=build /app/target/bin/notify.jar .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["java", "-jar", "notify.jar"]
