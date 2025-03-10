# iOS-CSV conversion tool

CLI tool for converting localization data between .csv and .lproj / localizable.strings format.

The purpose of the tool is to simplify the process of managing and converting localization data for iOS applications.

## Features

- Convert single .CSV file to multiple Localized.strings files (ex. sv.lproj/Localized.strings)
- Aggregate iOS localizations (eg. [en|sv|de|etc].lproj/Localizable.strings) to a single .CSV file

## Usage

To use the tool, run the following commands:

```bash
# Convert CSV data to iOS format
go run main.go -i=/path/to/input -o=/path/to/output

# Convert localization data to CSV format
go run main.go -i=/path/to/input -o=/path/to/output -tocsv
```

Ensure that your CSV file adheres to the following format:

```
Key,en,sv,de
hello,Hello,Hej,Hallo
```

The first column describes what language the column is intended for.

## Contributing

Contributions are welcome! Please submit a pull request with your changes.

## License

This project is licensed under the [MIT License](https://opensource.org/licenses/MIT).
