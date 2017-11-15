#include <atomic>
#include <chrono>
#include <iostream>
#include <map>
#include <sstream>
#include <thread>

#include <libcouchbase/api3.h>
#include <libcouchbase/couchbase.h>

using atomic_ptr_t = std::shared_ptr<std::atomic<int64_t>>;
// Used for checkpointing of vbucket seq nos
typedef std::map<int64_t, atomic_ptr_t> vb_seq_map_t;

struct Result {
  lcb_CAS cas;
  lcb_error_t rc;
  std::string value;
  uint32_t exptime;

  Result() : cas(0), rc(LCB_SUCCESS) {}
};

template <typename... Args>
std::string string_sprintf(const char *format, Args... args) {
  int length = std::snprintf(nullptr, 0, format, args...);
  assert(length >= 0);

  char *buf = new char[length + 1];
  std::snprintf(buf, length + 1, format, args...);

  std::string str(buf);
  delete[] buf;
  return str;
}

void get_callback(lcb_t instance, int cbtype, const lcb_RESPBASE *rb) {
  // lcb_get calls against metadata bucket is only triggered for timer lookups
  const lcb_RESPGET *rg = reinterpret_cast<const lcb_RESPGET *>(rb);
  const void *data = lcb_get_cookie(instance);

  std::string ts;
  std::string timestamp_marker("");
  lcb_CMDSTORE acmd = {0};
  Result res;

  switch (rb->rc) {
  case LCB_KEY_ENOENT:
    ts.assign(reinterpret_cast<const char *>(data));

    LCB_CMD_SET_KEY(&acmd, ts.c_str(), ts.length());
    LCB_CMD_SET_VALUE(&acmd, timestamp_marker.c_str(),
                      timestamp_marker.length());
    acmd.operation = LCB_ADD;

    lcb_store3(instance, &res, &acmd);
    lcb_wait(instance);
    break;
  case LCB_SUCCESS:
    std::cout << string_sprintf("Value %.*s", static_cast<int>(rg->nvalue),
                                reinterpret_cast<const char *>(rg->value));
    break;
  default:
    std::cout << "LCB_CALLBACK_GET: Operation failed, "
              << lcb_strerror(NULL, rb->rc) << " rc:" << rb->rc << '\n';
    break;
  }
}

void set_callback(lcb_t instance, int cbtype, const lcb_RESPBASE *rb) {
  const lcb_RESPSTORE *rs = reinterpret_cast<const lcb_RESPSTORE *>(rb);
  Result *result = reinterpret_cast<Result *>(rb->cookie);
  result->rc = rs->rc;
}

void sdmutate_callback(lcb_t instance, int cbtype, const lcb_RESPBASE *rb) {
  const lcb_RESPSUBDOC *resp = reinterpret_cast<const lcb_RESPSUBDOC *>(rb);
  lcb_SDENTRY ent;
  size_t iter = 0;

  Result *res = reinterpret_cast<Result *>(rb->cookie);
  res->rc = rb->rc;

  if (lcb_sdresult_next(resp, &ent, &iter)) {
    std::cout << string_sprintf("Status: 0x%x. Value: %.*s\n", ent.status,
                                static_cast<int>(ent.nvalue),
                                reinterpret_cast<const char *>(ent.value));
  }
}

void sdlookup_callback(lcb_t instance, int cbtype, const lcb_RESPBASE *rb) {
  Result *res = reinterpret_cast<Result *>(rb->cookie);
  res->cas = rb->cas;
  res->rc = rb->rc;

  if (rb->rc == LCB_SUCCESS) {
    const lcb_RESPGET *rg = reinterpret_cast<const lcb_RESPGET *>(rb);
    res->value.assign(reinterpret_cast<const char *>(rg->value), rg->nvalue);

    const lcb_RESPSUBDOC *resp = reinterpret_cast<const lcb_RESPSUBDOC *>(rb);
    lcb_SDENTRY ent;
    size_t iter = 0;
    int index = 0;
    while (lcb_sdresult_next(resp, &ent, &iter)) {
      std::cout << string_sprintf("Status: 0x%x. Value: %.*s\n", ent.status,
                                  static_cast<int>(ent.nvalue),
                                  reinterpret_cast<const char *>(ent.value));

      if (index == 0) {
        std::string exptime(reinterpret_cast<const char *>(ent.value));
        exptime.substr(0, static_cast<int>(ent.nvalue));

        unsigned long long int ttl;
        char *pEnd;
        ttl = strtoull(exptime.c_str(), &pEnd, 10);
        res->exptime = (uint32_t)ttl;
      }

      if (index == 1) {
        res->value.assign(reinterpret_cast<const char *>(ent.value),
                          static_cast<int>(ent.nvalue));
      }
      index++;
    }
  }
}

int main(){
  std::string meta_connstr;
  meta_connstr = "couchbase://localhost:12000/"
                 "eventing?username=eventing&select_bucket=true";

  lcb_t meta_cb_instance;

  lcb_create_st crst;
  memset(&crst, 0, sizeof crst);

  crst.version = 3;
  crst.v.v3.connstr = meta_connstr.c_str();
  crst.v.v3.type = LCB_TYPE_BUCKET;
  crst.v.v3.passwd = "asdasd";

  lcb_create(&meta_cb_instance, &crst);
  lcb_connect(meta_cb_instance);
  lcb_wait(meta_cb_instance);

  lcb_install_callback3(meta_cb_instance, LCB_CALLBACK_GET, get_callback);
  lcb_install_callback3(meta_cb_instance, LCB_CALLBACK_STORE, set_callback);
  lcb_install_callback3(meta_cb_instance, LCB_CALLBACK_SDMUTATE,
                        sdmutate_callback);
  lcb_install_callback3(meta_cb_instance, LCB_CALLBACK_SDLOOKUP,
                        sdlookup_callback);

  std::string appName("credit_score");
  vb_seq_map_t vb_seq;

  for (int i = 0; i < 1024; i++) {
    vb_seq[i] = atomic_ptr_t(new std::atomic<int64_t>(0));
  }

  const auto checkpoint_interval =
      std::chrono::milliseconds(1000);
  const auto sleep_duration = std::chrono::milliseconds(100);
  std::string seq_no_path("last_processed_seq_no");

  for (int i = 0; i < 1024; i++) {
    lcb_CMDSTORE scmd = { 0 };
    std::stringstream key, value;
    key << appName << "_vb_" << i;
    LCB_CMD_SET_KEY(&scmd, key.str().c_str(), key.str().length());
    value << "{\"version\": 5.1}";
    LCB_CMD_SET_VALUE(&scmd, value.str().c_str(), value.str().length());
    scmd.operation = LCB_SET;

    Result res;
    lcb_store3(meta_cb_instance, &res, &scmd);
    if (res.rc != LCB_SUCCESS) {
        std::cout << "Initial create rc: " << lcb_strerror(NULL, res.rc);
    }
  }

  while (true) {
    for (int i = 0; i < 1024; i++) {
      auto seq = vb_seq[i].get()->load(std::memory_order_seq_cst);
      if (seq == 0) {
        std::cout << "Processing vb: " << i << std::endl;
        std::stringstream vb_key;
        vb_key << appName << "_vb_" << i;

        lcb_CMDSUBDOC cmd = {0};
        LCB_CMD_SET_KEY(&cmd, vb_key.str().c_str(), vb_key.str().length());

        lcb_SDSPEC seq_spec = {0};
        seq_spec.sdcmd = LCB_SDCMD_DICT_UPSERT;
        seq_spec.options = LCB_SDSPEC_F_MKINTERMEDIATES;

        LCB_SDSPEC_SET_PATH(&seq_spec, seq_no_path.c_str(),
                            seq_no_path.length());
        auto seq_str = std::to_string(seq);
        LCB_SDSPEC_SET_VALUE(&seq_spec, seq_str.c_str(), seq_str.length());

        cmd.specs = &seq_spec;
        cmd.nspecs = 1;

        Result cres;
        int retry_count = 0;
        lcb_subdoc3(meta_cb_instance, &cres, &cmd);
        lcb_wait(meta_cb_instance);
        while (cres.rc != LCB_SUCCESS) {
          std::cout << __LINE__ << " vb: " << i
                    << " rc: " << lcb_strerror(NULL, cres.rc)
                    << " retry_count: " << retry_count << std::endl;
          retry_count++;
          std::this_thread::sleep_for(sleep_duration);
          lcb_subdoc3(meta_cb_instance, &cres, &cmd);
          lcb_wait(meta_cb_instance);
        }

        // Reset the seq no of checkpointed vb to 0
        if (cres.rc == LCB_SUCCESS) {
          vb_seq[i].get()->compare_exchange_strong(seq, 0);
        }
      }
    }
    std::this_thread::sleep_for(checkpoint_interval);
  }
}

