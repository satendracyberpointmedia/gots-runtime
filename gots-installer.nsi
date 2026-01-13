; GoTS Runtime - NSIS Installer Script
; Lightweight installer alternative to MSI

!define PRODUCT_NAME "GoTS Runtime"
!define PRODUCT_VERSION "0.1.0"
!define PRODUCT_PUBLISHER "GoTS Team"
!define PRODUCT_WEB_SITE "https://github.com/yourusername/gots-runtime"
!define PRODUCT_DIR_REGKEY "Software\Microsoft\Windows\CurrentVersion\App Paths\gots.exe"
!define PRODUCT_UNINST_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}"

; Compression
SetCompressor /SOLID lzma

; Modern UI
!include "MUI2.nsh"

; Required for environment variable manipulation
!include "EnvVarUpdate.nsh"

; General Settings
Name "${PRODUCT_NAME} ${PRODUCT_VERSION}"
OutFile "dist\windows\gots-runtime-${PRODUCT_VERSION}-windows-amd64-setup.exe"
InstallDir "$PROGRAMFILES64\GoTSRuntime"
InstallDirRegKey HKLM "${PRODUCT_DIR_REGKEY}" ""
ShowInstDetails show
ShowUnInstDetails show

; Request admin privileges
RequestExecutionLevel admin

; Interface Settings
!define MUI_ABORTWARNING
!define MUI_ICON "assets\gots.ico"
!define MUI_UNICON "assets\gots.ico"
!define MUI_WELCOMEFINISHPAGE_BITMAP "assets\installer-welcome.bmp"
!define MUI_HEADERIMAGE
!define MUI_HEADERIMAGE_BITMAP "assets\installer-header.bmp"

; Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES

; Finish page options
!define MUI_FINISHPAGE_RUN "$INSTDIR\gots.exe"
!define MUI_FINISHPAGE_RUN_PARAMETERS "--version"
!define MUI_FINISHPAGE_RUN_TEXT "Run 'gots --version' to verify installation"
!define MUI_FINISHPAGE_SHOWREADME "$INSTDIR\README.md"
!define MUI_FINISHPAGE_SHOWREADME_TEXT "View README"
!insertmacro MUI_PAGE_FINISH

; Uninstaller pages
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; Language
!insertmacro MUI_LANGUAGE "English"

; Version Information
VIProductVersion "${PRODUCT_VERSION}.0"
VIAddVersionKey "ProductName" "${PRODUCT_NAME}"
VIAddVersionKey "ProductVersion" "${PRODUCT_VERSION}"
VIAddVersionKey "FileVersion" "${PRODUCT_VERSION}"
VIAddVersionKey "FileDescription" "Go-based TypeScript Runtime"
VIAddVersionKey "LegalCopyright" "© ${PRODUCT_PUBLISHER}"
VIAddVersionKey "CompanyName" "${PRODUCT_PUBLISHER}"

; Installation Section
Section "Core Runtime" SecCore
  SectionIn RO  ; Required section
  
  SetOutPath "$INSTDIR"
  
  ; Main files
  File "build\windows\gots.exe"
  File "README.md"
  
  ; Copy LICENSE if exists
  IfFileExists "LICENSE" 0 +2
    File "LICENSE"
  
  ; Copy entire stdlib directory
  SetOutPath "$INSTDIR\stdlib"
  File /r "build\windows\stdlib\*.*"
  
  SetOutPath "$INSTDIR"
  
  ; Create uninstaller
  WriteUninstaller "$INSTDIR\uninstall.exe"
  
  ; Registry keys
  WriteRegStr HKLM "${PRODUCT_DIR_REGKEY}" "" "$INSTDIR\gots.exe"
  WriteRegStr HKLM "${PRODUCT_UNINST_KEY}" "DisplayName" "${PRODUCT_NAME}"
  WriteRegStr HKLM "${PRODUCT_UNINST_KEY}" "UninstallString" "$INSTDIR\uninstall.exe"
  WriteRegStr HKLM "${PRODUCT_UNINST_KEY}" "DisplayIcon" "$INSTDIR\gots.exe"
  WriteRegStr HKLM "${PRODUCT_UNINST_KEY}" "DisplayVersion" "${PRODUCT_VERSION}"
  WriteRegStr HKLM "${PRODUCT_UNINST_KEY}" "Publisher" "${PRODUCT_PUBLISHER}"
  WriteRegStr HKLM "${PRODUCT_UNINST_KEY}" "URLInfoAbout" "${PRODUCT_WEB_SITE}"
  
  ; Estimate size
  ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
  IntFmt $0 "0x%08X" $0
  WriteRegDWORD HKLM "${PRODUCT_UNINST_KEY}" "EstimatedSize" "$0"
  
SectionEnd

Section "Add to PATH" SecPath
  ; Add installation directory to system PATH
  ${EnvVarUpdate} $0 "PATH" "A" "HKLM" "$INSTDIR"
  
  ; Set GOTS_STDLIB_PATH environment variable
  WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "GOTS_STDLIB_PATH" "$INSTDIR\stdlib"
  
  ; Broadcast environment change
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
  
SectionEnd

Section "Start Menu Shortcuts" SecShortcuts
  CreateDirectory "$SMPROGRAMS\${PRODUCT_NAME}"
  CreateShortcut "$SMPROGRAMS\${PRODUCT_NAME}\GoTS Runtime.lnk" "$INSTDIR\gots.exe" "--help"
  CreateShortcut "$SMPROGRAMS\${PRODUCT_NAME}\README.lnk" "$INSTDIR\README.md"
  CreateShortcut "$SMPROGRAMS\${PRODUCT_NAME}\Uninstall.lnk" "$INSTDIR\uninstall.exe"
SectionEnd

; Section descriptions
!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${SecCore} "Core GoTS runtime and standard library (required)"
  !insertmacro MUI_DESCRIPTION_TEXT ${SecPath} "Add GoTS to system PATH for easy access from any directory"
  !insertmacro MUI_DESCRIPTION_TEXT ${SecShortcuts} "Create Start Menu shortcuts"
!insertmacro MUI_FUNCTION_DESCRIPTION_END

; Uninstaller Section
Section "Uninstall"
  ; Remove files
  Delete "$INSTDIR\gots.exe"
  Delete "$INSTDIR\README.md"
  Delete "$INSTDIR\LICENSE"
  Delete "$INSTDIR\uninstall.exe"
  
  ; Remove stdlib directory
  RMDir /r "$INSTDIR\stdlib"
  
  ; Remove shortcuts
  Delete "$SMPROGRAMS\${PRODUCT_NAME}\*.*"
  RMDir "$SMPROGRAMS\${PRODUCT_NAME}"
  
  ; Remove from PATH
  ${un.EnvVarUpdate} $0 "PATH" "R" "HKLM" "$INSTDIR"
  
  ; Remove environment variable
  DeleteRegValue HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "GOTS_STDLIB_PATH"
  
  ; Remove registry keys
  DeleteRegKey HKLM "${PRODUCT_UNINST_KEY}"
  DeleteRegKey HKLM "${PRODUCT_DIR_REGKEY}"
  
  ; Remove installation directory
  RMDir "$INSTDIR"
  
  ; Broadcast environment change
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
  
SectionEnd

; Initialize function - check if already installed
Function .onInit
  ReadRegStr $R0 HKLM "${PRODUCT_UNINST_KEY}" "UninstallString"
  StrCmp $R0 "" done
  
  MessageBox MB_OKCANCEL|MB_ICONEXCLAMATION \
  "${PRODUCT_NAME} is already installed.$\n$\nClick 'OK' to remove the previous version or 'Cancel' to cancel this upgrade." \
  IDOK uninst
  Abort
  
uninst:
  ClearErrors
  ExecWait '$R0 _?=$INSTDIR'
  
done:
FunctionEnd