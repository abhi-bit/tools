#include <libcouchbase/api3.h>
#include <libcouchbase/couchbase.h>
#include <iostream>

template <typename... Args>
std::string string_sprintf(const char *format, Args... args) {
  int length = std::snprintf(nullptr, 0, format, args...);
  assert(length >= 0);

  char *buf = new char[length + 1];
  std::snprintf(buf, length + 1, format, args...);

  std::string str(buf);
  delete[] buf;
  return std::move(str);
}

static void multi_op_callback(lcb_t instance, int cbtype,
                              const lcb_RESPBASE *rb) {
  std::cout << "Got callback for " << lcb_strcbtype(cbtype) << '\n';

  // if (rb->rc != LCB_SUCCESS && rb->rc != LCB_SUBDOC_MULTI_FAILURE) {
  //   std::cout << "Operation failed, " << lcb_strerror(NULL, rb->rc)
  //             << " rc:" << rb->rc << '\n';
  //   return;
  // }

  if (cbtype == LCB_CALLBACK_GET) {
    const lcb_RESPGET *rg = reinterpret_cast<const lcb_RESPGET *>(rb);

    const void *data = lcb_get_cookie(instance);
    std::string ts;
    std::string timestamp_marker("");
    lcb_CMDSTORE acmd = {0};

    switch (rb->rc) {
      case LCB_KEY_ENOENT:
        ts.assign((const char *)data);
        // std::cout << (const char *)data << '\n';

        LCB_CMD_SET_KEY(&acmd, ts.c_str(), ts.length());
        LCB_CMD_SET_VALUE(&acmd, timestamp_marker.c_str(),
                          timestamp_marker.length());
        acmd.operation = LCB_ADD;

        lcb_sched_enter(instance);
        lcb_store3(instance, NULL, &acmd);
        lcb_sched_leave(instance);
        lcb_wait(instance);
        break;

      case LCB_SUCCESS:
        std::cout << string_sprintf("Value %.*s", (int)rg->nvalue, rg->value)
                  << '\n';
        break;
      default:
        std::cout << "Operation failed, " << lcb_strerror(NULL, rb->rc)
                  << " rc:" << rb->rc << '\n';
        break;
    }

  } else if (cbtype == LCB_CALLBACK_SDMUTATE ||
             cbtype == LCB_CALLBACK_SDLOOKUP) {
    const lcb_RESPSUBDOC *resp = reinterpret_cast<const lcb_RESPSUBDOC *>(rb);
    lcb_SDENTRY ent;
    size_t iter = 0;
    if (lcb_sdresult_next(resp, &ent, &iter)) {
      std::cout << string_sprintf("Status: 0x%x. Value: %.*s\n", ent.status,
                                  (int)ent.nvalue, ent.value)
                << '\n';
    }
  }
}

int main() {
  std::string rbac_pass;
  rbac_pass.assign("asdasd");
  std::string connstr =
      "couchbase://localhost:12000?username=eventing&select_bucket=true";

  lcb_create_st crst;
  memset(&crst, 0, sizeof crst);

  crst.version = 3;
  crst.v.v3.connstr = connstr.c_str();
  crst.v.v3.type = LCB_TYPE_BUCKET;
  crst.v.v3.passwd = rbac_pass.c_str();

  lcb_t cb_instance;

  lcb_create(&cb_instance, &crst);
  lcb_connect(cb_instance);
  lcb_wait(cb_instance);

  lcb_install_callback3(cb_instance, LCB_CALLBACK_DEFAULT, multi_op_callback);

  std::string func, ts, val;

  func.assign("nonDocTimer");
  ts.assign("2017-06-05T20:00:00");

  val.assign(";{\"callback_func\": \"");
  val.append(func);
  val.append("\", \"start_ts\": \"");
  val.append(ts);
  val.append("\"}");

  lcb_CMDGET gcmd = {0};
  LCB_CMD_SET_KEY(&gcmd, ts.c_str(), ts.length());

  lcb_get3(cb_instance, NULL, &gcmd);
  lcb_set_cookie(cb_instance, ts.c_str());
  lcb_wait(cb_instance);
  lcb_set_cookie(cb_instance, NULL);

  // appending delimiter ";"
  lcb_CMDSTORE cmd = { 0 };
  cmd.operation = LCB_APPEND;

  LCB_CMD_SET_KEY(&cmd, ts.c_str(), ts.length());
  LCB_CMD_SET_VALUE(&cmd, val.c_str(), val.length());
  lcb_sched_enter(cb_instance);
  lcb_store3(cb_instance, NULL, &cmd);
  lcb_sched_leave(cb_instance);
  lcb_wait(cb_instance);
}
