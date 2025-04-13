param(
    [string]$version = "1.0.0"
)

# Construct the output file name using the version.
$binaryName = "gotion_v$version.exe"

Write-Host "Building gotion version $version..."
Write-Host "Output binary: $binaryName"

# Run the build command.
# Note: On Windows and in PowerShell, you can pass ldflags without inner quotes.
go build -ldflags "-X main.version=$version" -o $binaryName

if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed!"
} else {
    Write-Host "Build succeeded."
}