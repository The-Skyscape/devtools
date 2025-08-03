## MIT License - Copyright (c) 2025 TheSkyscape

#########################################################
###                                                   ###
###    TheSkyscape DevTools Makefile                  ###
###      Building the tools for cloud development     ###
###                                                   ###
###              .                                    ###
###              |				                      ###
###     .               /			                  ###
###      \       I     /                              ###
###        \  ,g88RR_                                 ###
###          d888(`  ).                               ###
### -  --==  888(     ).=--           .+(`  )`        ###
###)         Y8P(       '`.          :(   .    )      ###
###        .+(`(      .   )     .--  `.  (    ) )     ###
###       ((    (..__.:'-'   .=(   )   ` _`  ) )      ###
###`.     `(       ) )       (   .  )     (   )  .__  ###
###  )      ` __.:'   )     (   (   ))     `-'.:(`   )###
###)  )  ( )       --'       `- __.'         :(       ###
###.-'  (_.'          .')                    `(     ( ###
###                  (_  )                     ` __.:'###
###                                                   ###
### --..,___.--,--'`,---..-.--+--.,,-,,..._.--..-._.-_###
#########################################################

TOOLS := create-app launch-app

.PHONY: all clean install

all: $(addprefix build/,$(TOOLS))

clean:
	rm -rf build

artifacts:
	@mkdir -p ./build
	@touch ./build/create-app
	@touch ./build/launch-app

install: artifacts
	go install ./cmd/create-app
	go install ./cmd/launch-app

build/create-app: artifacts
	go build -o $@ ./cmd/create-app

build/launch-app: artifacts
	go build -o $@ ./cmd/launch-app