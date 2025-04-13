# gotion
a cli tool for uploading json and csv records to notion databases 

<br>

# Build

```
go build -o gotion.exe
```

<br>

# Testing

```
gotion insert --db "your-database-id" --data "data.json" --api-key "your-api-key"
```

### Testing Data
```json
[
  {
    "properties": {
      "Name": "Testing",
      "desc": "Sample description",
      "Age": 25
    }
  },
  {
    "properties": {
      "Name": "Testing",
      "desc": "Sample description",
      "Age": 25
    }
  }
]
```