# Output:
NO_COLOR		?= false
ifeq ($(NO_COLOR),false)
ECHO				:= echo -e
C_STD 			= $(shell $(ECHO) -e "\033[0m")
C_RED				= $(shell $(ECHO) -e "\033[91m")
C_GREEN 		= $(shell $(ECHO) -e "\033[92m")
C_YELLOW		= $(shell $(ECHO) -e "\033[93m")
C_BLUE			= $(shell $(ECHO) -e "\033[94m")
I_CROSS 		= $(shell $(ECHO) -e "\xe2\x95\xb3")
I_CHECK 		= $(shell $(ECHO) -e "\xe2\x9c\x94")
I_BULLET		= $(shell $(ECHO) -e "\xe2\x80\xa2")
else
ECHO				:= echo
C_STD 			=
C_RED				=
C_GREEN 		=
C_YELLOW		=
C_BLUE			=
I_CROSS 		= x
I_CHECK 		= .
I_BULLET		= *
endif
