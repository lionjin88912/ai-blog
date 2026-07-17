; AI Blog 部落格 — Windows installer (NSIS)
;
; Per-user install (no admin / UAC prompt — corporate friendly):
;   Program : %LOCALAPPDATA%\Programs\ai-blog\ai-blog.exe
;   Data    : %LOCALAPPDATA%\ai-blog\  (created by the app itself; the
;             uninstaller leaves it alone unless the user opts in)
; Build:  makensis -DVERSION=1.0.0 installer.nsi
; Inputs: ai-blog.exe (the windows build) + ai-blog.ico in this directory.

Unicode true
!ifndef VERSION
  !define VERSION "0.0.0"
!endif

!define APPNAME "AI Blog 部落格"
!define SLUG "ai-blog"
!define UNINSTKEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${SLUG}"

Name "${APPNAME}"
OutFile "ai-blog-setup.exe"
RequestExecutionLevel user
InstallDir "$LOCALAPPDATA\Programs\${SLUG}"
Icon "ai-blog.ico"
UninstallIcon "ai-blog.ico"

!include "MUI2.nsh"
!define MUI_ICON "ai-blog.ico"
!define MUI_UNICON "ai-blog.ico"

!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!define MUI_FINISHPAGE_RUN "$INSTDIR\ai-blog.exe"
!define MUI_FINISHPAGE_RUN_TEXT "立即啟動 ${APPNAME}"
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "TradChinese"

Section "Install"
  ; If AI Blog is running its exe is locked and can't be overwritten. Remind
  ; the user (no NSIS process plugin, so this is a prompt rather than a kill),
  ; and use "try" so one locked file doesn't abort the whole install.
  IfFileExists "$INSTDIR\ai-blog.exe" 0 fresh
    MessageBox MB_OKCANCEL|MB_ICONEXCLAMATION \
      "偵測到已安裝的 AI Blog。$\n請先關閉正在執行的 AI Blog(含瀏覽器分頁)再繼續,否則檔案可能無法更新。" \
      IDOK fresh
      Abort "已取消:請關閉 AI Blog 後重試。"
  fresh:
  SetOverwrite try
  SetOutPath "$INSTDIR"
  File "ai-blog.exe"
  File "ai-blog.ico"

  ; Shortcuts (Start Menu + Desktop)
  CreateShortcut "$SMPROGRAMS\${APPNAME}.lnk" "$INSTDIR\ai-blog.exe" "" "$INSTDIR\ai-blog.ico"
  CreateShortcut "$DESKTOP\${APPNAME}.lnk" "$INSTDIR\ai-blog.exe" "" "$INSTDIR\ai-blog.ico"

  ; Uninstaller + Add/Remove Programs entry (per-user, HKCU)
  WriteUninstaller "$INSTDIR\uninstall.exe"
  WriteRegStr HKCU "${UNINSTKEY}" "DisplayName" "${APPNAME}"
  WriteRegStr HKCU "${UNINSTKEY}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKCU "${UNINSTKEY}" "DisplayIcon" "$INSTDIR\ai-blog.ico"
  WriteRegStr HKCU "${UNINSTKEY}" "Publisher" "Lion Travel"
  WriteRegStr HKCU "${UNINSTKEY}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
  WriteRegDWORD HKCU "${UNINSTKEY}" "NoModify" 1
  WriteRegDWORD HKCU "${UNINSTKEY}" "NoRepair" 1
SectionEnd

Section "Uninstall"
  Delete "$SMPROGRAMS\${APPNAME}.lnk"
  Delete "$DESKTOP\${APPNAME}.lnk"
  Delete "$INSTDIR\ai-blog.exe"
  Delete "$INSTDIR\ai-blog.ico"
  Delete "$INSTDIR\uninstall.exe"
  RMDir "$INSTDIR"
  DeleteRegKey HKCU "${UNINSTKEY}"

  ; Data (personas, credentials, downloaded tools) is precious — ask before
  ; removing, default is to KEEP it so a reinstall picks everything back up.
  MessageBox MB_YESNO|MB_ICONQUESTION|MB_DEFBUTTON2 \
    "也要刪除所有資料嗎?$\n(人格、WordPress 帳密、發文紀錄、已下載工具)$\n$\n選「否」保留資料,之後重裝可直接接續使用。" \
    IDNO keepdata
    RMDir /r "$LOCALAPPDATA\${SLUG}"
  keepdata:
SectionEnd
