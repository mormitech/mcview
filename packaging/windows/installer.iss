; Inno Setup script for mcView. Per-user install (no admin rights), adds a Start menu entry, no startup entry.
; Silent install, as winget runs it:  mcview-setup-<version>-windows-x64.exe /VERYSILENT /SUPPRESSMSGBOXES /NORESTART
; CI replaces __VERSION__ and compiles it (see .github/workflows/build.yml).
[Setup]
AppId={{FBB0D79E-25C4-4781-8EDC-DC5CB39DD1AD}
AppName=mcView
AppVersion=__VERSION__
AppPublisher=mormitech
AppPublisherURL=https://github.com/mormitech/mcview
AppSupportURL=https://github.com/mormitech/mcview/issues
DefaultDirName={autopf}\mcview
DisableDirPage=yes
DisableProgramGroupPage=yes
DisableReadyPage=yes
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
OutputDir=..\..\dist
OutputBaseFilename=mcview-setup-__VERSION__-windows-x64
UninstallDisplayName=mcView
UninstallDisplayIcon={app}\mcview.exe
Compression=lzma2
SolidCompression=yes
; tells Explorer about the new file types, so that Open with... offers mcView at once
ChangesAssociations=yes

[Files]
Source: "..\..\build\bin\mcview.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\LICENSE"; DestDir: "{app}"; Flags: ignoreversion

; Open with... Per user (HKCU), no admin rights. Windows does not let an installer make a program the default for a file type:
; mcView is offered in Open with... and in Settings > Apps > Default apps, and the user picks it there. The list of types is the
; one in image.go (a test keeps them equal).
[Registry]
Root: HKCU; Subkey: "Software\Classes\mcview.Picture"; ValueType: string; ValueData: "mcView picture"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\Classes\mcview.Picture\DefaultIcon"; ValueType: string; ValueData: "{app}\mcview.exe,0"
Root: HKCU; Subkey: "Software\Classes\mcview.Picture\shell\open\command"; ValueType: string; ValueData: """{app}\mcview.exe"" ""%1"""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "mcView"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\shell\open\command"; ValueType: string; ValueData: """{app}\mcview.exe"" ""%1"""
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities"; ValueType: string; ValueName: "ApplicationName"; ValueData: "mcView"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities"; ValueType: string; ValueName: "ApplicationDescription"; ValueData: "Picture viewer"
Root: HKCU; Subkey: "Software\RegisteredApplications"; ValueType: string; ValueName: "mcView"; ValueData: "Software\Mormitech\mcView\Capabilities"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.jpg\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.jpeg\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.jpe\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.jfif\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.png\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.gif\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.bmp\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.webp\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.ico\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.svg\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.avif\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.tif\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.tiff\OpenWithProgids"; ValueType: none; ValueName: "mcview.Picture"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".jpg"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".jpeg"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".jpe"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".jfif"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".png"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".gif"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".bmp"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".webp"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".ico"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".svg"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".avif"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".tif"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe\SupportedTypes"; ValueType: string; ValueName: ".tiff"; ValueData: ""
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".jpg"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".jpeg"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".jpe"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".jfif"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".png"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".gif"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".bmp"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".webp"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".ico"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".svg"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".avif"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".tif"; ValueData: "mcview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mcView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".tiff"; ValueData: "mcview.Picture"

[Icons]
Name: "{autoprograms}\mcView"; Filename: "{app}\mcview.exe"

[Run]
Filename: "{app}\mcview.exe"; Description: "Start mcView"; Flags: nowait postinstall skipifsilent

[Code]
// taskkill returns before Windows has released the file of the killed process, so wait a moment before the installer
// replaces or deletes it.
procedure StopMcprint;
var
  ResultCode: Integer;
begin
  Exec(ExpandConstant('{sys}\taskkill.exe'), '/F /IM mcview.exe', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Sleep(1500);
end;

// An upgrade replaces a file that may be in use, so close the running app first.
procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssInstall then StopMcprint;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usUninstall then StopMcprint;
end;
