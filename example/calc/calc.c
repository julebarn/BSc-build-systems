
#include "numbers.h"
#include <stdio.h>

int main() {
    int a = 5;
    int b = 3;

    int sum = add(a, b);
    int difference = sub(a, b);
    int product = mult(a, b);

    printf("Sum: %d\n", sum);
    printf("Difference: %d\n", difference);
    printf("Product: %d\n", product);

    return 0;
}