# Ajoute le bin Google Cloud SDK au PATH utilisateur
$binPath = "C:\Users\fabio\AppData\Local\Google\Cloud SDK\google-cloud-sdk\bin"
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*google-cloud-sdk*") {
  [Environment]::SetEnvironmentVariable("Path", $userPath + ";" + $binPath, "User")
  Write-Host "OK: Chemin ajoute au PATH utilisateur: $binPath"
  Write-Host "Ferme et rouvre ton terminal (ou Cursor) pour que ce soit pris en compte."
} else {
  Write-Host "Le chemin Google Cloud SDK est deja dans le PATH utilisateur."
}
