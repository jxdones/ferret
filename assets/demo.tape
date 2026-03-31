Output assets/ferret.gif

Require echo

Set Shell "zsh"
Set FontSize 14
Set Width 1300
Set Height 800
Set WindowBar Rings
Set Padding 10
Set BorderRadius 10

Type "ferret"
Sleep 600ms
Enter

Wait+Screen /Ready/
Sleep 500ms

Ctrl+U
Sleep 500ms
Type "https://httpbin.org/delay/1"
Sleep 200ms
Escape
Sleep 250ms
Ctrl+R
Sleep 1s

Wait+Screen /{/
Sleep 600ms
Type "]"
Sleep 600ms
Type "]"
Sleep 600ms
Type "]"
Sleep 600ms
Type "]"
Sleep 600ms

Type "T"
Sleep 500ms
Escape
Sleep 500ms
Type "c"
Sleep 400ms
Type "/"
Sleep 400ms

Type "get users"
Sleep 500ms
Enter
Sleep 400ms
Ctrl+R
Sleep 500ms
Wait+Screen /{/
Sleep 500ms

Type "T"
Sleep 500ms
Escape
Sleep 500ms
Type "cc"
Sleep 300ms
Type "/"
Sleep 300ms

Type "get ditto"
Sleep 500ms
Enter
Sleep 400ms
Ctrl+R
Sleep 500ms
Wait+Screen /{/
Sleep 500ms

Type "X"
Sleep 400ms
Type "X"
Sleep 400ms

Type "T"
Sleep 500ms
Escape
Sleep 500ms
Type "c"
Sleep 400ms
Type "/"
Sleep 400ms

Type "create post"
Sleep 500ms
Enter
Sleep 400ms
Shift+Tab
Sleep 400ms
Type "]"
Sleep 600ms
Type "]"
Sleep 600ms
Ctrl+R
Sleep 500ms
Wait+Screen /{/
Sleep 500ms

Type "X"
Sleep 500ms

Escape
Sleep 800ms
Type "?"
Sleep 1s
Escape

Sleep 2s
