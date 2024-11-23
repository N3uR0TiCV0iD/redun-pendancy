@echo off
    SETLOCAL EnableDelayedExpansion
    echo go mod tidy
    go mod tidy
    echo.
    echo go build -x .
    echo.
    go build -x .
    echo.
    echo DONE^^!
    if "%1" NEQ "no-exit" (
        echo.
        echo Press any key to exit. . .
        pause>nul
    )
goto :EOF
