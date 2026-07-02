; LanShare NSIS Installer for Windows
; Requires: NSIS (https://nsis.sourceforge.io/)

!define PRODUCT_NAME "LanShare"
!define PRODUCT_VERSION "1.0.0"
!define PRODUCT_PUBLISHER "gensui-fuga"
!define PRODUCT_WEB_SITE "https://github.com/gensui-fuga/lanshare"

SetCompressor lzma

; Modern UI
!include "MUI2.nsh"

!define MUI_ABORTWARNING
!define MUI_ICON "${NSISDIR}\Contrib\Graphics\Icons\modern-install.ico"

; Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; Languages
!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "SimpChinese"

Name "${PRODUCT_NAME} ${PRODUCT_VERSION}"
OutFile "dist/windows/LanShare-${PRODUCT_VERSION}-Setup.exe"
InstallDir "$PROGRAMFILES64\LanShare"
ShowInstDetails show

Section "Install"
    SetOutPath "$INSTDIR"
    
    ; Main executable
    File "dist/windows/lanshare.exe"
    
    ; Create start menu shortcut
    CreateDirectory "$SMPROGRAMS\LanShare"
    CreateShortCut "$SMPROGRAMS\LanShare\LanShare.lnk" "$INSTDIR\lanshare.exe"
    CreateShortCut "$SMPROGRAMS\LanShare\Uninstall.lnk" "$INSTDIR\uninstall.exe"
    
    ; Desktop shortcut
    CreateShortCut "$DESKTOP\LanShare.lnk" "$INSTDIR\lanshare.exe"
    
    ; Write uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"
    
    ; Registry for Add/Remove Programs
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" \
        "DisplayName" "${PRODUCT_NAME}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" \
        "UninstallString" "$INSTDIR\uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" \
        "DisplayVersion" "${PRODUCT_VERSION}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}" \
        "Publisher" "${PRODUCT_PUBLISHER}"
SectionEnd

Section "Uninstall"
    ; Remove files
    Delete "$INSTDIR\lanshare.exe"
    Delete "$INSTDIR\uninstall.exe"
    RMDir "$INSTDIR"
    
    ; Remove shortcuts
    Delete "$SMPROGRAMS\LanShare\*"
    RMDir "$SMPROGRAMS\LanShare"
    Delete "$DESKTOP\LanShare.lnk"
    
    ; Remove registry
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}"
SectionEnd
