package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jomei/notionapi"
	"github.com/spf13/cobra"
)

// RawPageData represents the raw JSON data before converting to Notion API format
type RawPageData map[string]interface{}

// PageData structure for final Notion API compatible data
type PageData struct {
	Properties notionapi.Properties `json:"properties"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gotion",
		Short: "Gotion is a CLI tool to upload data to Notion tables",
		Long: `Gotion is a CLI tool to upload data to Notion tables.
Database IDs should be in UUID format like: "f1a2b3c4-d5e6-7f8a-9b0c-1d2e3f4a5b6c"
These can be found in the Notion URL when viewing a database.`,
	}

	// Add inspect command to show database schema
	var inspectCmd = &cobra.Command{
		Use:   "inspect",
		Short: "Inspect a Notion database to see its structure",
		Long: `Inspect a Notion database to see its structure.
This helps ensure your JSON data will map correctly to the database columns.

Example: gotion inspect --db "f1a2b3c4-d5e6-7f8a-9b0c-1d2e3f4a5b6c"`,
		Run: func(cmd *cobra.Command, args []string) {
			dbID, _ := cmd.Flags().GetString("db")
			apiKeyFlag, _ := cmd.Flags().GetString("api-key")
			
			// --- Input Validation --- 
			if dbID == "" {
				fmt.Println("Error: Database ID (--db) is required.")
				os.Exit(1)
			}

			// Validate UUID format for database ID
			dbID = cleanDatabaseID(dbID)
			if !isValidUUID(dbID) {
				fmt.Println("Error: The database ID must be in UUID format.")
				fmt.Println("Example: f1a2b3c4-d5e6-7f8a-9b0c-1d2e3f4a5b6c")
				fmt.Println("You can find this in your Notion URL when viewing the database.")
				os.Exit(1)
			}

			// --- Get API Key --- 
			apiKey := apiKeyFlag
			if apiKey == "" {
				apiKey = os.Getenv("NOTION_API_KEY")
			}

			if apiKey == "" {
				fmt.Println("Error: Notion API key not provided. Set via --api-key flag or NOTION_API_KEY environment variable.")
				os.Exit(1)
			}

			// --- Initialize Notion Client --- 
			client := notionapi.NewClient(notionapi.Token(apiKey))
			ctx := context.Background()

			// Get database info
			fmt.Printf("Inspecting database %s...\n", dbID)
			database, err := client.Database.Get(ctx, notionapi.DatabaseID(dbID))
			if err != nil {
				fmt.Printf("Error accessing database: %v\n", err)
				
				if strings.Contains(err.Error(), "Could not find database") {
					fmt.Println("\nPermission Error: Your integration doesn't have access to this database.")
					fmt.Println("To fix this:")
					fmt.Println("1. Go to your database in Notion")
					fmt.Println("2. Click the \"...\" menu in the top right corner")
					fmt.Println("3. Select \"Add connections\"")
					fmt.Println("4. Find and select your integration name")
					fmt.Println("\nAlso verify that your Database ID is correct.")
				}
				
				os.Exit(1)
			}

			// Display database info
			fmt.Printf("\nDatabase Title: %s\n", getTitle(database.Title))
			fmt.Printf("\nProperties (columns) available:\n")
			fmt.Println("-----------------------------")
			
			for name, property := range database.Properties {
				fmt.Printf("%s (Type: %s)\n", name, getPropertyTypeString(property))
			}
			
			fmt.Println("\nWhen creating your JSON data file, make sure property names exactly match these column names.")
			fmt.Println("Example for this database:")
			fmt.Println("```")
			fmt.Println(`{
  "properties": {`)
			
			// Generate a sample property JSON for each property type
			for name, property := range database.Properties {
				fmt.Printf("    \"%s\": %s,\n", name, getSamplePropertyJSON(property))
			}
			
			fmt.Println(`  }
}`)
			fmt.Println("```")
		},
	}

	var insertCmd = &cobra.Command{
		Use:   "insert",
		Short: "Insert data from a JSON file into a Notion database",
		Long: `Insert data from a JSON file into a Notion database.
		
Example: gotion insert --db "f1a2b3c4-d5e6-7f8a-9b0c-1d2e3f4a5b6c" --data "data.json"

Note: The database ID should be a valid Notion UUID, found in the URL 
when viewing your database in Notion. For example:

In URL: https://www.notion.so/1d4a5e7fe23180b98df2ddce1ea05ddf?v=...
The database ID is: 1d4a5e7fe23180b98df2ddce1ea05ddf

IMPORTANT: You must share your database with your integration for this to work:
1. Go to your database in Notion
2. Click the "..." menu in the top right corner
3. Select "Add connections" 
4. Find and select your integration name

You can use the database ID with or without dashes. The tool will format it correctly.`,
		Run: func(cmd *cobra.Command, args []string) {
			dbID, _ := cmd.Flags().GetString("db")
			dataFile, _ := cmd.Flags().GetString("data")
			apiKeyFlag, _ := cmd.Flags().GetString("api-key")
			debugMode, _ := cmd.Flags().GetBool("debug")

			// --- Input Validation --- 
			if dbID == "" || dataFile == "" {
				fmt.Println("Error: Both --db (Database ID) and --data (JSON file path) flags are required.")
				os.Exit(1)
			}

			// Validate UUID format for database ID
			dbID = cleanDatabaseID(dbID)
			if !isValidUUID(dbID) {
				fmt.Println("Error: The database ID must be in UUID format.")
				fmt.Println("Example: f1a2b3c4-d5e6-7f8a-9b0c-1d2e3f4a5b6c")
				fmt.Println("You can find this in your Notion URL when viewing the database.")
				os.Exit(1)
			}

			// --- Get API Key --- 
			apiKey := apiKeyFlag
			if apiKey == "" {
				apiKey = os.Getenv("NOTION_API_KEY")
			}

			if apiKey == "" {
				fmt.Println("Error: Notion API key not provided. Set via --api-key flag or NOTION_API_KEY environment variable.")
				os.Exit(1)
			}

			// --- Initialize Notion Client --- 
			client := notionapi.NewClient(notionapi.Token(apiKey))
			ctx := context.Background()

			// --- Read Data File --- 
			fmt.Printf("Reading data from %s...\n", dataFile)
			content, err := os.ReadFile(dataFile)
			if err != nil {
				fmt.Printf("Error reading data file %s: %v\n", dataFile, err)
				os.Exit(1)
			}

			if debugMode {
				fmt.Println("Raw JSON content:")
				fmt.Println(string(content))
			}

			// --- Parse Raw JSON Data First --- 
			var rawData []RawPageData
			err = json.Unmarshal(content, &rawData)
			if err != nil {
				// Try as single object if array fails
				var singleRawData RawPageData
				errSingle := json.Unmarshal(content, &singleRawData)
				if errSingle != nil {
					fmt.Printf("Error parsing JSON data: %v\n", err)
					fmt.Println("Make sure your JSON is valid and contains property data for Notion.")
					os.Exit(1)
				}
				rawData = []RawPageData{singleRawData}
			}

			fmt.Printf("Found %d record(s) to insert into database %s.\n", len(rawData), dbID)

			// --- Convert Raw Data to Notion API Format --- 
			pagesData := make([]PageData, 0, len(rawData))

			// Fetch database schema
			database, err := client.Database.Get(ctx, notionapi.DatabaseID(dbID))
			if err != nil {
				fmt.Printf("Error accessing database: %v\n", err)
				os.Exit(1)
			}

			for i, raw := range rawData {
				if debugMode {
					fmt.Printf("Processing record %d...\n", i + 1)
				}

				// Convert raw data to Notion properties
				pageData, err := convertToNotionProperties(raw, *database)
				if err != nil {
					fmt.Printf("Error converting record %d: %v\n", i + 1, err)
					continue
				}

				pagesData = append(pagesData, pageData)
			}

			// --- Insert Data into Notion --- 
			successCount := 0
			for i, pageData := range pagesData {
				fmt.Printf("Inserting record %d... ", i + 1)
				
				// Validate property names against database schema if debug mode is on
				if debugMode {
					// First check if we can fetch the database schema
					database, err := client.Database.Get(ctx, notionapi.DatabaseID(dbID))
					if err == nil {
						// Check if the properties in our data match the database schema
						for propName := range pageData.Properties {
							if _, exists := database.Properties[propName]; !exists {
								fmt.Printf("\nWARNING: Property '%s' does not exist in the database schema!\n", propName)
								fmt.Printf("Available properties are: ")
								for dbPropName := range database.Properties {
									fmt.Printf("%s, ", dbPropName)
								}
								fmt.Println("\nEnsure your property names match exactly (including case).")
							}
						}
					}
				}
				
				request := &notionapi.PageCreateRequest{
					Parent: notionapi.Parent{
						DatabaseID: notionapi.DatabaseID(dbID),
					},
					Properties: pageData.Properties,
				}

				if debugMode {
					requestJSON, _ := json.MarshalIndent(request, "", "  ")
					fmt.Printf("\nRequest JSON:\n%s\n", string(requestJSON))
				}

				response, err := client.Page.Create(ctx, request)
				if err != nil {
					fmt.Printf("Failed: %v\n", err)
					
					// Check for common permission errors
					if strings.Contains(err.Error(), "Could not find database") {
						fmt.Println("\nPermission Error: Your integration doesn't have access to this database.")
						fmt.Println("To fix this:")
						fmt.Println("1. Go to your database in Notion")
						fmt.Println("2. Click the \"...\" menu in the top right corner")
						fmt.Println("3. Select \"Add connections\"")
						fmt.Println("4. Find and select your integration name")
						fmt.Println("\nAlso verify that your Database ID is correct.")
						
						// Only show this detailed help for the first error
						if i == 0 {
							fmt.Println("\nFor more help, visit: https://developers.notion.com/docs/getting-started")
						}
					}
				} else {
					fmt.Println("Success!")
					successCount++
					
					// Print URL of created page if available
					if response.URL != "" {
						fmt.Printf("Page URL: %s\n", response.URL)
					}
				}
			}
			
			if successCount > 0 {
				fmt.Printf("\nFinished inserting. %d/%d records inserted successfully.\n", successCount, len(pagesData))
				fmt.Println("\nTIP: If your data isn't visible in Notion:")
				fmt.Println("1. Verify property names match exactly with database columns (case sensitive)")
				fmt.Println("2. Run 'gotion inspect --db \"your-db-id\"' to see the database structure")
				fmt.Println("3. Try running with --debug flag to see more details about the process")
			} else {
				fmt.Printf("\nFinished inserting. %d/%d records inserted successfully.\n", successCount, len(pagesData))
				fmt.Println("No records were successfully inserted. Check the errors above.")
			}
		},
	}

	insertCmd.Flags().String("db", "", "ID of the Notion database")
	insertCmd.Flags().String("data", "", "Path to the JSON file")
	insertCmd.Flags().String("api-key", "", "Notion API Key (optional, overrides NOTION_API_KEY env var)")
	insertCmd.Flags().Bool("debug", false, "Enable debug mode for verbose output")
	insertCmd.MarkFlagRequired("db")
	insertCmd.MarkFlagRequired("data")

	// Add common flags
	inspectCmd.Flags().String("db", "", "ID of the Notion database")
	inspectCmd.Flags().String("api-key", "", "Notion API Key (optional, overrides NOTION_API_KEY env var)")
	inspectCmd.MarkFlagRequired("db")
	
	// Add commands to root
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(insertCmd)

	rootCmd.Execute()
}

// isValidUUID checks if the input string is a valid UUID
func isValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

// cleanDatabaseID attempts to extract a UUID from various formats
// (like URLs or when dashes are missing)
func cleanDatabaseID(input string) string {
	// If it contains a dash already, it might be a proper UUID format
	if strings.Contains(input, "-") {
		return input
	}
	
	// Check if it's a 32-character hex string without dashes
	r := regexp.MustCompile("^[a-fA-F0-9]{32}$")
	if r.MatchString(input) {
		// Insert dashes in the UUID format positions
		return fmt.Sprintf("%s-%s-%s-%s-%s",
			input[0:8],
			input[8:12],
			input[12:16],
			input[16:20],
			input[20:32])
	}
	
	// Extract ID from URL if it appears to be a Notion URL
	if strings.Contains(input, "notion.so") {
		parts := strings.Split(input, "/")
		lastPart := parts[len(parts)-1]
		// Check if the last part might be an ID
		if len(lastPart) >= 32 {
			// Try to clean this last part
			return cleanDatabaseID(lastPart)
		}
	}
	
	// Return as is if we can't determine a better format
	return input
}

// Dynamically handle all property types based on the database schema
func convertToNotionProperties(raw RawPageData, schema notionapi.Database) (PageData, error) {
    var result PageData
    result.Properties = make(notionapi.Properties)

    // Check if raw has a "properties" key
    if props, ok := raw["properties"].(map[string]interface{}); ok {
        for propName, propValue := range props {
            schemaProp, exists := schema.Properties[propName]
            if !exists {
                continue // Skip properties not in the database schema
            }
            var notionProp notionapi.Property
            switch schemaProp.GetType() {
            case notionapi.PropertyConfigTypeTitle:
                // If needed, you can pass the value as-is or transform further
                if strValue, ok := propValue.(string); ok {
                    notionProp = &notionapi.TitleProperty{
                        Title: []notionapi.RichText{
                            {
                                Text: &notionapi.Text{
                                    Content: strValue,
                                },
                            },
                        },
                    }
                }
            case notionapi.PropertyConfigTypeRichText:
                if strValue, ok := propValue.(string); ok {
                    notionProp = &notionapi.RichTextProperty{
                        RichText: []notionapi.RichText{
                            {
                                Text: &notionapi.Text{
                                    Content: strValue,
                                },
                            },
                        },
                    }
                }
            case notionapi.PropertyConfigTypeNumber:
                if numValue, ok := propValue.(float64); ok {
                    notionProp = &notionapi.NumberProperty{
                        Number: numValue,
                    }
                }
            default:
                continue // Skip unsupported property types
            }
            
            if notionProp != nil {
                result.Properties[propName] = notionProp
            }
        }
    } else {
        // Optionally handle the case where raw is not structured with a "properties" key.
        return result, fmt.Errorf("expected key 'properties' in data, got none")
    }
    
    return result, nil
}

// Helper functions for database inspection
func getTitle(titleArray []notionapi.RichText) string {
	if len(titleArray) == 0 {
		return "Untitled"
	}
	
	var title string
	for _, text := range titleArray {
		if text.Text != nil {
			title += text.Text.Content
		}
	}
	
	return title
}

func getPropertyTypeString(prop notionapi.PropertyConfig) string {
	switch prop.GetType() {
	case notionapi.PropertyConfigTypeTitle:
		return "Title"
	case notionapi.PropertyConfigTypeRichText:
		return "Rich Text"
	case notionapi.PropertyConfigTypeNumber:
		return "Number"
	case notionapi.PropertyConfigTypeSelect:
		return "Select"
	case notionapi.PropertyConfigTypeMultiSelect:
		return "Multi Select"
	case notionapi.PropertyConfigTypeDate:
		return "Date"
	case notionapi.PropertyConfigTypePeople:
		return "People"
	case notionapi.PropertyConfigTypeFiles:
		return "Files"
	case notionapi.PropertyConfigTypeCheckbox:
		return "Checkbox"
	case notionapi.PropertyConfigTypeURL:
		return "URL"
	case notionapi.PropertyConfigTypeEmail:
		return "Email"
	case notionapi.PropertyConfigTypePhoneNumber:
		return "Phone Number"
	case notionapi.PropertyConfigTypeFormula:
		return "Formula"
	case notionapi.PropertyConfigTypeRelation:
		return "Relation"
	case notionapi.PropertyConfigTypeRollup:
		return "Rollup"
	// Remove or comment out the undefined constants
	// case notionapi.PropertyConfigTypeCreatedTime:
	//	return "Created Time"
	// case notionapi.PropertyConfigTypeCreatedBy:
	//	return "Created By"
	// case notionapi.PropertyConfigTypeLastEditedTime:
	//	return "Last Edited Time"
	// case notionapi.PropertyConfigTypeLastEditedBy:
	//	return "Last Edited By"
	default:
		// Handle type as string for unknown/newer property types
		return string(prop.GetType())
	}
}

func getSamplePropertyJSON(prop notionapi.PropertyConfig) string {
	switch prop.GetType() {
	case notionapi.PropertyConfigTypeTitle:
		return `{"title": [{"text": {"content": "Sample Title"}}]}`
	case notionapi.PropertyConfigTypeRichText:
		return `{"rich_text": [{"text": {"content": "Sample text"}}]}`
	case notionapi.PropertyConfigTypeNumber:
		return `{"number": 42}`
	case notionapi.PropertyConfigTypeSelect:
		return `{"select": {"name": "Option Name"}}`
	case notionapi.PropertyConfigTypeMultiSelect:
		return `{"multi_select": [{"name": "Option 1"}, {"name": "Option 2"}]}`
	case notionapi.PropertyConfigTypeDate:
		return `{"date": {"start": "2023-01-01", "end": null}}`
	case notionapi.PropertyConfigTypeCheckbox:
		return `{"checkbox": true}`
	case notionapi.PropertyConfigTypeURL:
		return `{"url": "https://example.com"}`
	case notionapi.PropertyConfigTypeEmail:
		return `{"email": "example@example.com"}`
	case notionapi.PropertyConfigTypePhoneNumber:
		return `{"phone_number": "+1 234 567 8900"`
	default:
		return `{/* Complex type - see Notion API docs */}`
	}
}