#include "queue.h"

#include <iostream>

#define SIZE 10

int main() {
    auto queue = new Queue<int>();

    for (int i = 0; i < SIZE; i++) {
      queue->push(i);
    }

    std::cout << "queue size: " << queue->count() << std::endl;

    auto count = queue->count();

    for (int i = 0; i < count; i++) {
        std::cout << queue->pop() << std::endl;
    }

}
