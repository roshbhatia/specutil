Require specutil

Set Shell "zsh"
Set Theme "Catppuccin Macchiato"
Set FontSize 16
Set Width 1400
Set Height 760
Set Padding 20
Set Framerate 20
Set TypingSpeed 30ms

Sleep 1s

Type "cd ~/rotation-service"
Enter
Sleep 1s

Type "specutil graph --as mermaid"
Enter
Sleep 3s

Type "specutil next rotate-signing-keys"
Enter
Sleep 4s
