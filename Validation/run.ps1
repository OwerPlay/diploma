param(
    [string]$DashboardPath = 'C:/git/AlektaDashboard',
    [string]$PrinterPath = 'C:/git/AlektaPrinter'
)
$ErrorActionPreference = 'Stop'
$validationRoot = Join-Path ([IO.Path]::GetTempPath()) ('alekta-validation-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $validationRoot | Out-Null
$projects = @(
    @{ Name='dashboard'; Path=$DashboardPath; Revision='9e39ed795bcb110bb8de252ac44f677c3d8e4de7' },
    @{ Name='printer'; Path=$PrinterPath; Revision='57b079a4c0c8847965fcc475a98675cf87a3e145' }
)
foreach ($project in $projects) {
    $archive = Join-Path $validationRoot ($project.Name + '.zip')
    $destination = Join-Path $validationRoot $project.Name
    git -C $project.Path archive $project.Revision -o $archive
    if ($LASTEXITCODE -ne 0) { throw 'Unable to archive the documented revision.' }
    Expand-Archive -LiteralPath $archive -DestinationPath $destination
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot ($project.Name + '_test.go')) -Destination (Join-Path $destination 'diploma_test.go')
    Push-Location $destination
    try {
        go test -v ./...
        if ($LASTEXITCODE -ne 0) { throw ('Validation failed: ' + $project.Name) }
    } finally { Pop-Location }
}
Write-Output ('Validation copies: ' + $validationRoot)
