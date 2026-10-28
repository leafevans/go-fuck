@rem go-fuck shim for cmd.exe, modeled after thefuck's scripts/fuck.bat.
@rem
@rem It feeds the previous command from the doskey history to go-fuck and
@rem executes the corrected command in the current console. go-fuck must be on
@rem PATH, and PowerShell must be available (it ships with Windows).
@rem
@rem For a persistent setup prefer the PowerShell shim scripts/fuck.ps1 or add
@rem "iex \"$(go-fuck --alias)\"" to your PowerShell profile.
@powershell -NoProfile -Command "cmd /c \"$(go-fuck %* $(doskey /history)[-2])\"; [Console]::ResetColor();"
