#include <iostream>
#include <fstream>

#include "./include/v8_init_generated.h"

int main(int argc, char *argv[]) {
  std::ifstream ifs(argv[1]);
  std::string content(std::istreambuf_iterator<char>(ifs),
                     (std::istreambuf_iterator<char>()));

  auto cfg = v8init::GetInit((const void *)content.c_str());

  std::cout << "appname: " << cfg->appname()->str() << " khp: "
      << cfg->kvhostport()->str() << " depcfg: " << cfg->depcfg()->str() << std::endl;
}
