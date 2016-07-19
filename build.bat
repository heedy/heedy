@echo off
:: This is JUST the build command. You must
:: run dependencies.bat first to set up the folders

:: https://stackoverflow.com/questions/6359820/how-to-set-commands-output-as-a-variable-in-a-batch-file
FOR /F "tokens=* USEBACKQ" %%F IN (`git rev-parse HEAD`) DO (
SET gitsha=%%F
)
ECHO git sha: %gitsha%

:: https://stackoverflow.com/questions/203090/how-to-get-current-datetime-on-windows-command-line-in-a-suitable-format-for-us
For /f "tokens=2-4 delims=/ " %%a in ('date /t') do (set mydate=%%c-%%a-%%b)
For /f "tokens=1-2 delims=/:" %%a in ("%TIME%") do (set mytime=%%a%%b)
SET curtime=%mydate%
ECHO build date: %curtime%
:: TODO: The curtime command should be %mydate%_%mytime%, but mytime seems to have a space, which makes go unhappy
:: Also... should make it utc.

go build -o bin/connectordb.exe  -ldflags "-X commands.BuildStamp=%curtime% -X commands.GitHash=%gitsha%" src/main.go

exit /B
