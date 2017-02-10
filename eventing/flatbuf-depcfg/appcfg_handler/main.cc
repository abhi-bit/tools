#include <iostream>
#include <fstream>

#include "./include/depcfg_schema_generated.h"

int main(int argc, char *argv[]) {
  std::ifstream ifs(argv[1]);
  std::string content(std::istreambuf_iterator<char>(ifs),
                     (std::istreambuf_iterator<char>()));

  auto cfg = appcfg::GetConfig((const void *)content.c_str());

  std::cout << "id: " << cfg->id() << " code: " << cfg->appCode()->str()
            << " name: " << cfg->appName()->str() << std::endl;

  auto dd = cfg->depCfg();
  std::cout << "tick duration: " << dd->tickDuration() << " wc: " << dd->workerCount()
      << " auth: " << dd->auth()->str() << " meta: " << dd->metadataBucket()->str()
      << " src: " << dd->sourceBucket()->str() << std::endl;

  auto bdd = dd->buckets();
  for (unsigned int i = 0; i < bdd->size(); i++) {
      std::cout << "alias: " << bdd->Get(i)->alias()->str() << " bucketname: "
          << bdd->Get(i)->bucketName()->str() << std::endl;
  }
  return 0;
}
