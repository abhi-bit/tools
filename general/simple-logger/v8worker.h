#include "./log.h"

class Foo {
public:
    Foo(int a);
    ~Foo();
    void PrintA();
private:
    int a;
};

Foo::Foo(int entry) {
    a = entry;
}

Foo::~Foo() {
    a = -1;
}

void Foo::PrintA() {
    LOG(logDebug) << "value of a => " << a << '\n';
}
