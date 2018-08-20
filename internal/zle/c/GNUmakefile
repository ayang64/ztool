SOURCES	:= $(wildcard *.c)
OBJECTS	:= $(SOURCES:.c=.o)

CFLAGS	+= -MMD -std=gnu99 -I/usr/src/sys/cddl/contrib/opensolaris/uts/common/

.PHONY : clean

zle-test : $(OBJECTS)
	$(CC) $(OBJECTS) -o zle-test

$(OBJECTS) : $(SOURCES)

-include *.d

clean:
	rm $(OBJECTS) *.d zle-test
