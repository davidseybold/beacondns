# Beacon DNS Client

This is a Go client library for interacting with the Beacon DNS API. It provides a simple interface for managing DNS zones and resource record sets.

## Installation

```bash
go get github.com/davidseybold/beacondns/client
```

## Usage

### Creating a Client

```go
import "github.com/davidseybold/beacondns/client"

// Create a new client with the base URL of your Beacon DNS API
c := client.New("http://localhost:8080")
```

### Managing Zones

#### Create a Zone

```go
ctx := context.Background()
response, err := c.CreateZone(ctx, "example.com")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created zone: %s\n", response.Zone.ID)
```

#### List Zones

```go
response, err := c.ListZones(ctx)
if err != nil {
    log.Fatal(err)
}
for _, zone := range response.Zones {
    fmt.Printf("Zone: %s (%s)\n", zone.Name, zone.ID)
}
```

#### Get Zone Information

```go
zone, err := c.GetZone(ctx, "zone-id")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Zone: %s\n", zone.Name)
```

### Managing Resource Record Sets

#### List Resource Record Sets

```go
response, err := c.ListResourceRecordSets(ctx, "zone-id")
if err != nil {
    log.Fatal(err)
}
for _, rrset := range response.ResourceRecordSets {
    fmt.Printf("Record: %s %s\n", rrset.Name, rrset.Type)
}
```

#### Change Resource Record Sets

```go
changes := []client.Change{
    {
        Action: "CREATE",
        ResourceRecordSet: client.ResourceRecordSet{
            Name: "www.example.com",
            Type: "A",
            TTL:  300,
            ResourceRecords: []client.ResourceRecord{
                {Value: "192.0.2.1"},
            },
        },
    },
}

changeInfo, err := c.ChangeResourceRecordSets(ctx, "zone-id", changes)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Change submitted: %s\n", changeInfo.ID)
```

## Error Handling

The client returns errors in a consistent format. All API errors include a code and message:

```go
response, err := c.CreateZone(ctx, "example.com")
if err != nil {
    // Error will be in the format: "API error: ErrorCode - Error message"
    log.Fatal(err)
}
```

## Context Support

All client methods accept a context.Context parameter, allowing you to control timeouts and cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

response, err := c.ListZones(ctx)
if err != nil {
    log.Fatal(err)
}
``` 