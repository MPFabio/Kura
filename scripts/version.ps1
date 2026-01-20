# Script PowerShell de gestion des versions pour ModulOps
# Usage: .\scripts\version.ps1 [major|minor|patch] ou .\scripts\version.ps1 [version] (ex: 1.2.3)

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("major", "minor", "patch")]
    [string]$Type,
    
    [Parameter(Mandatory=$false)]
    [string]$Version
)

$VERSION_FILE = "VERSION"
$CHANGELOG_FILE = "CHANGELOG.md"

# Lire la version actuelle
if (Test-Path $VERSION_FILE) {
    $CURRENT_VERSION = Get-Content $VERSION_FILE -Raw | ForEach-Object { $_.Trim() }
} else {
    $CURRENT_VERSION = "0.0.0"
}

# Fonction pour incrémenter la version
function Increment-Version {
    param(
        [string]$CurrentVersion,
        [string]$IncrementType
    )
    
    $parts = $CurrentVersion -split '\.'
    $major = [int]$parts[0]
    $minor = [int]$parts[1]
    $patch = [int]$parts[2]
    
    switch ($IncrementType) {
        "major" {
            $major++
            $minor = 0
            $patch = 0
        }
        "minor" {
            $minor++
            $patch = 0
        }
        "patch" {
            $patch++
        }
    }
    
    return "$major.$minor.$patch"
}

# Déterminer la nouvelle version
if ($Version) {
    # Vérifier le format X.Y.Z
    if ($Version -match '^\d+\.\d+\.\d+$') {
        $NEW_VERSION = $Version
    } else {
        Write-Host "Format de version invalide. Utilisez X.Y.Z (ex: 1.2.3)" -ForegroundColor Red
        exit 1
    }
} elseif ($Type) {
    $NEW_VERSION = Increment-Version -CurrentVersion $CURRENT_VERSION -IncrementType $Type
} else {
    Write-Host "Version actuelle: $CURRENT_VERSION"
    Write-Host "Usage: .\scripts\version.ps1 -Type [major|minor|patch] ou .\scripts\version.ps1 -Version [version]"
    exit 1
}

Write-Host "Version actuelle: $CURRENT_VERSION" -ForegroundColor Cyan
Write-Host "Nouvelle version: $NEW_VERSION" -ForegroundColor Green

# Mettre à jour le fichier VERSION
$NEW_VERSION | Out-File -FilePath $VERSION_FILE -Encoding utf8 -NoNewline
Write-Host "✓ Fichier VERSION mis à jour" -ForegroundColor Green

# Mettre à jour package.json du frontend
if (Test-Path "frontend\package.json") {
    $packageJson = Get-Content "frontend\package.json" -Raw | ConvertFrom-Json
    $packageJson.version = $NEW_VERSION
    $packageJson | ConvertTo-Json -Depth 10 | Set-Content "frontend\package.json"
    Write-Host "✓ frontend/package.json mis à jour" -ForegroundColor Green
}

# Créer un tag Git
$response = Read-Host "Créer un tag Git v$NEW_VERSION ? (y/n)"
if ($response -eq 'y' -or $response -eq 'Y') {
    git add $VERSION_FILE, $CHANGELOG_FILE
    if (Test-Path "frontend\package.json") {
        git add frontend/package.json
    }
    git commit -m "chore: bump version to $NEW_VERSION" 2>&1 | Out-Null
    git tag -a "v$NEW_VERSION" -m "Version $NEW_VERSION"
    Write-Host "✓ Tag Git v$NEW_VERSION créé" -ForegroundColor Green
    Write-Host ""
    Write-Host "Pour pousser le tag vers le dépôt distant:" -ForegroundColor Yellow
    Write-Host "  git push origin v$NEW_VERSION"
    Write-Host "  git push origin main  # ou votre branche"
}

Write-Host ""
Write-Host "Version $NEW_VERSION prête !" -ForegroundColor Green
Write-Host "N'oubliez pas de mettre à jour CHANGELOG.md avec les changements de cette version." -ForegroundColor Yellow
