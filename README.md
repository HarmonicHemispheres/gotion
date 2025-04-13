![](gotion_banner.png)

<!-- Shield Icons -->
<p align="center">
    <a href="https://github.com/HarmonicHemispheres/gotion">
        <img src="https://img.shields.io/github/repo-size/HarmonicHemispheres/gotion.svg" alt="Repository Size" />
    </a>
    <a href="https://github.com/HarmonicHemispheres/gotion/commits/main">
        <img src="https://img.shields.io/github/last-commit/HarmonicHemispheres/gotion.svg" alt="Last Commit" />
    </a>
    <img alt="GitHub Downloads (all assets, all releases)" src="https://img.shields.io/github/downloads/HarmonicHemispheres/gotion/total">
    <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/HarmonicHemispheres/gotion">


  <br>
  <i>
a cli tool for uploading json and csv records to notion databases  
  </i>
</p>
<center>
</center>


<br>

# Testing

Pre-Req:
- make sure to create a custom integration on your notion workspace
- enable the integration on your notion database

### Testing Data
```json
[
  {
    "properties": {
      "Name": "Jellybean Report",
      "desc": "An experimental dataset about candy preferences.",
      "Age": 42,
      "Website": "https://sweetdata.io",
      "Select": "REC",
      "TestDate": { "date": { "start": "2025-03-22T00:00:00-07:00", "end": null } },
      "Multi": ["Option A", "Option B", "Option C"],
      "Checkbox": true,
      "Contact": "example@example.com",
      "Phone": "+1 234 567 8900"
    }
  },
  {
    "properties": {
      "Name": "Testing2",
      "desc": "Sample description",
      "Age": 25,
      "Website": "https://google.com",
      "Select": "C",
      "Multi": ["Option A", "Option B", "Option C"],
      "TestDate": { "date": { "start": "2025-05-12T00:00:00-07:00", "end": null } }
    }
  }
]
```
### Insert data into Notion


```
gotion insert --db "your-database-id" --data "data.json" --api-key "your-api-key"
```


![alt text](demo_1.png)


<br>
<br>

# CLI Commands

### `gotion inspect`

Inspect a Notion database to see its structure, including property names and types. This is useful for ensuring your JSON data matches the database schema.

```
gotion inspect --db "your-database-id" --api-key "your-api-key"
```

### `gotion insert`

Insert data from a JSON file into a specified Notion database.

```
gotion insert --db "your-database-id" --data "path/to/your/data.json" --api-key "your-api-key"
```

**Flags:**
*   `--db`: (Required) The ID of the Notion database.
*   `--data`: (Required) The path to the JSON file containing the data to insert.
*   `--api-key`: (Optional) Your Notion API key. If not provided, it will check for the `NOTION_API_KEY` environment variable.
*   `--debug`: (Optional) Enable debug mode for more verbose output, including the request payload sent to Notion.

### `gotion version`

Prints the current version of the gotion CLI tool.

```
gotion version
```

<br>
<br>

# Building from Source

```ps1
.\make.ps1 -version "1.2.3"
```
