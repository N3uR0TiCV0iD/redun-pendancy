@echo off
    SETLOCAL EnableDelayedExpansion
    call :CheckStaleDependencies
    if !_returnValue!==1 (
        echo Stale dependencies detected^^!
        echo.
        echo Building project. . .
        echo.
        call build.bat no-exit
    )
    echo.
    echo go run .
    echo.
    go run .
    echo.
    echo Press any key to exit. . .
    pause>nul
goto :EOF

:CheckStaleDependencies
    echo Checking for stale dependencies. . .
    FOR /F "tokens=1,2 delims=:" %%A in ('go list -f "{{.Stale}}:{{.ImportPath}}" -tags=all all') DO (
        if "%%A"=="true" (
            call :strbegins "%%B" "redun-pendancy"
            if !_returnValue!==0 (
                set _returnValue=1
                goto :EOF
            )
        )
    )
    set _returnValue=0
goto :EOF

::https://stackoverflow.com/a/5841587
::strlen(string)
:strlen
    (set^ __currString=%~1)
    if NOT defined __currString (
        set _returnValue=0
        goto :EOF
    )
    set _returnValue=1
    FOR %%P IN (4096 2048 1024 512 256 128 64 32 16 8 4 2 1) DO (
        if NOT "!__currString:~%%P,1!"=="" (
            set /a _returnValue+=%%P
            set __currString=!__currString:~%%P!
        )
    )
goto :EOF

::strbegins(text, value)
:strbegins
    set __text=%~1
    call :strlen %2
    set __substring=!__text:~0,%_returnValue%!
    if "!__substring!"=="%~2" (
        set _returnValue=1
        goto :EOF
    )
    set _returnValue=0
goto :EOF
