# SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
#
# SPDX-License-Identifier: CC0-1.0

Output screenshots/demo.webm
Require go
Require echo
Set Theme "Catppuccin Latte"
Set FontFamily "Berkeley Mono"
Set MarginFill "#00897b"
Set Margin 25
Set BorderRadius 10
Set WindowBar ColorfulRight
Set WindowBarSize 60
Set TypingSpeed 25ms
Set Shell "zsh"
Set FontSize 32
Set Width 1500
Set Height 800

Hide
Type "export DIR=$(mktemp -d) && clear"
Enter
Type "CGO_ENABLED=0 go build -ldflags='-s -w' -o willow ./cmd && clear"
Enter
Type "mv willow $DIR/ && clear"
Enter
Type "cd $DIR && clear"
Enter
Show

Type "ls"
Sleep 500ms 
Enter
Type "./willow" 
Sleep 500ms 
Enter
Type "ls" 
Sleep 500ms 
Enter
Sleep 3s
Type "vim config.toml" 
Sleep 500ms 
Enter
Down@100ms 7
Type@100ms "wwwvE" 
Left 1 
Type@100ms "c0.0.0.0:3000" 
Sleep 500ms 
Escape 
Type@100ms ":wq" 
Enter
Type "./willow" 
Sleep 500ms 
Enter
Sleep 3s
Ctrl+C 
Sleep 500ms
Type "ls" 
Sleep 500ms 
Enter
Sleep 5s
Hide
Type "rm -rf $DIR && clear"
