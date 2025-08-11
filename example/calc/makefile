


calc: calc.o add.o sub.o mult.o
	gcc -o calc calc.o add.o sub.o mult.o

calc.o: calc.c numbers.h
	gcc -c calc.c

add.o: add.c numbers.h
	gcc -c add.c

sub.o: sub.c numbers.h
	gcc -c sub.c	

mult.o: mult.c numbers.h
	gcc -c mult.c	

clean:
	rm -f calc *.o
	