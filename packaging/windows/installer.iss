; Inno Setup script for mormiView. Per-user install (no admin rights), adds a Start menu entry, no startup entry.
; Silent install, as winget runs it:  mormiview-setup-<version>-windows-x64.exe /VERYSILENT /SUPPRESSMSGBOXES /NORESTART
; CI replaces __VERSION__ and compiles it (see .github/workflows/build.yml).
[Setup]
AppId={{FBB0D79E-25C4-4781-8EDC-DC5CB39DD1AD}
AppName=mormiView
AppVersion=__VERSION__
AppPublisher=mormitech
AppPublisherURL=https://github.com/mormitech/mormiview
AppSupportURL=https://github.com/mormitech/mormiview/issues
DefaultDirName={autopf}\mormiview
DisableDirPage=yes
DisableProgramGroupPage=yes
DisableReadyPage=yes
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
OutputDir=..\..\dist
OutputBaseFilename=mormiview-setup-__VERSION__-windows-x64
UninstallDisplayName=mormiView
UninstallDisplayIcon={app}\mormiview.exe
Compression=lzma2
SolidCompression=yes
; tells Explorer about the new file types, so that Open with... offers mormiView at once
ChangesAssociations=yes

; mcView was the name before: an install of it is upgraded (the AppId is the same), so its program, entries and registry keys are removed
[InstallDelete]
Type: files; Name: "{app}\mcview.exe"
Type: files; Name: "{userprograms}\mcView.lnk"

[Files]
Source: "..\..\build\bin\mormiview.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\LICENSE"; DestDir: "{app}"; Flags: ignoreversion

; Open with... Per user (HKCU), no admin rights. Windows does not let an installer make a program the default for a file type:
; mormiView is offered in Open with... and in Settings > Apps > Default apps, and the user picks it there. The list of types is the
; one in image.go (a test keeps them equal).
[Registry]
Root: HKCU; Subkey: "Software\Classes\mcview.Picture"; Flags: dontcreatekey deletekey
Root: HKCU; Subkey: "Software\Classes\Applications\mcview.exe"; Flags: dontcreatekey deletekey
Root: HKCU; Subkey: "Software\Mormitech\mcView"; Flags: dontcreatekey deletekey
Root: HKCU; Subkey: "Software\RegisteredApplications"; ValueName: "mcView"; Flags: dontcreatekey deletevalue
Root: HKCU; Subkey: "Software\Classes\mormiview.Picture"; ValueType: string; ValueData: "mormiView picture"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\Classes\mormiview.Picture\DefaultIcon"; ValueType: string; ValueData: "{app}\mormiview.exe,0"
Root: HKCU; Subkey: "Software\Classes\mormiview.Picture\shell\open\command"; ValueType: string; ValueData: """{app}\mormiview.exe"" ""%1"""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "mormiView"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\shell\open\command"; ValueType: string; ValueData: """{app}\mormiview.exe"" ""%1"""
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities"; ValueType: string; ValueName: "ApplicationName"; ValueData: "mormiView"; Flags: uninsdeletekey
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities"; ValueType: string; ValueName: "ApplicationDescription"; ValueData: "Picture viewer"
Root: HKCU; Subkey: "Software\RegisteredApplications"; ValueType: string; ValueName: "mormiView"; ValueData: "Software\Mormitech\mormiView\Capabilities"; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.jpg\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.jpeg\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.jpe\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.jfif\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.png\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.gif\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.bmp\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.webp\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.ico\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.svg\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.avif\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.tif\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\.tiff\OpenWithProgids"; ValueType: string; ValueName: "mormiview.Picture"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".jpg"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".jpeg"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".jpe"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".jfif"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".png"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".gif"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".bmp"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".webp"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".ico"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".svg"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".avif"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".tif"; ValueData: ""
Root: HKCU; Subkey: "Software\Classes\Applications\mormiview.exe\SupportedTypes"; ValueType: string; ValueName: ".tiff"; ValueData: ""
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".jpg"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".jpeg"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".jpe"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".jfif"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".png"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".gif"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".bmp"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".webp"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".ico"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".svg"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".avif"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".tif"; ValueData: "mormiview.Picture"
Root: HKCU; Subkey: "Software\Mormitech\mormiView\Capabilities\FileAssociations"; ValueType: string; ValueName: ".tiff"; ValueData: "mormiview.Picture"

[Icons]
Name: "{autoprograms}\mormiView"; Filename: "{app}\mormiview.exe"

[Run]
Filename: "{app}\mormiview.exe"; Description: "Start mormiView"; Flags: nowait postinstall skipifsilent

[Code]
// taskkill returns before Windows has released the file of the killed process, so wait a moment before the installer
// replaces or deletes it.
procedure StopMcprint;
var
  ResultCode: Integer;
begin
  Exec(ExpandConstant('{sys}\taskkill.exe'), '/F /IM mormiview.exe', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
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
