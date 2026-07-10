; OBS Client Full Installer Script
; Includes WebView2 Runtime installation (offline)

#define MyAppName "OBS Client"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "OBS Client"
#define MyAppExeName "obs-client.exe"
#define MyAppIcon "..\icon.ico"
#define SourceDir "..\..\bin"
#define WebView2Installer "MicrosoftEdgeWebView2RuntimeInstallerX64.exe"

[Setup]
AppId={{OBS-CLIENT-APP-GUID-FULL}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableDirPage=no
DisableProgramGroupPage=yes
OutputDir=..\..\bin
OutputBaseFilename=obs-client-setup-full-{#MyAppVersion}
SetupIconFile={#MyAppIcon}
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
UninstallDisplayIcon={app}\{#MyAppExeName}
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
ShowLanguageDialog=no
PrivilegesRequired=admin

[Tasks]
Name: "desktopicon"; Description: "Create desktop shortcut"; GroupDescription: "Create shortcuts:"; Flags: unchecked

[Files]
Source: "{#SourceDir}\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#WebView2Installer}"; DestDir: "{tmp}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Uninstall {#MyAppName}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{tmp}\{#WebView2Installer}"; Parameters: "/silent /install"; Description: "Installing WebView2 Runtime..."; Flags: waituntilterminated; StatusMsg: "Installing WebView2 Runtime..."
Filename: "{app}\{#MyAppExeName}"; Description: "Launch {#MyAppName}"; Flags: nowait postinstall skipifsilent

[UninstallDelete]
Type: filesandordirs; Name: "{app}"