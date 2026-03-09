
The goal is to build a TUI called rundown, that reads and displays a Markdown document in the terminal.
And it allows to easily execute codeblocks in the document, similar to what org bable / org tangle does.


## Tasks

1. [x] Finish the bootstrap of the project. We already ran agent-ops in it but the project itself is not yet a ready made go project. Make the quality gate pass
2. [x] Build the first version of the program that renders the output of the markdown and the outline in a two pane view


## Technical requirements

* Use golang bubbletea in version 2
* Use golang version 1.26
* Add a mise.toml to lockdown the required tools

## TUI layout and requirements

The TUI will follow the following layout:

| Header            |
+-------------------+
|             |     |
|             |     |
|             |     |
|             |     |
|             |     |
|             |     |
+-------------+-----+
|Footer             |
+-------------------+

So we have a header, a main sections with two panes and a footer.
The two pane section shows the rendered markdown on the left.
On the right it shows and outline of the sections and it also demarcates nodes that are executable separately.
It uses nerdfont icons to show the icon of the language to execute if it exists, otherwise a generic execute symbol.
The markdown view must be scrollable.
I have to be able to switch between markdown and outline with tab.
They are bound to each other meaning, that moving in the markdown panel and scrolling synchronizes with the outline.
Also navigating in the outline synchronizes with the markdown view. 

Key bindings:

* C-c or C-q or Q to quit

* In markdown pane
* hjkl for navigation 
* HJKL for navigation on heading level 
* In outline
* c for collapse (current)
* C for collapse all
* e for expand (current)
* E for expand all
* x to show only exectuable targets
* hj for navigation between headings
* n to go to next executable target 
* p to go to preview executable target
* r - nothing for now (will be used to run the target later) 
