@echo off
setlocal
echo ======================================================================
echo  FreeTab Pro - Standalone Build Script
echo ======================================================================
echo.

echo [1/2] Compiling FreeTab.exe for Windows (amd64)...
call wails build -s -platform windows/amd64
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Compilation failed!
    exit /b %ERRORLEVEL%
)

echo.
echo [2/2] Build Verification:
if exist "build\bin\FreeTab.exe" (
    echo [SUCCESS] FreeTab.exe generated successfully at:
    echo           build\bin\FreeTab.exe
    for %%I in ("build\bin\FreeTab.exe") do echo           File Size: %%~zI bytes
) else (
    echo [ERROR] build\bin\FreeTab.exe was not found!
    exit /b 1
)

echo.
echo ======================================================================
echo  Build complete! Ready for GitHub Releases distribution.
echo ======================================================================
endlocal
