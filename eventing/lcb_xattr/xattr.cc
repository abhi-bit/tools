#define LCB_NO_DEPR_CXX_CTORS
#undef NDEBUG

/*
 * Sample run
 * ➜  tmp ./test "couchbase://127.0.0.1:12000/default?select_bucket=true&username=eventing" asdasd
 * Storing the initial item..
 * Got callback for STORE.. OK
 * Got callback for SDMUTATE.. No result!
 * Retrieving 'ram'
 * ====
 * Got callback for GET.. Value {"hello":"world"}
 * ====
 * */

#include <string>
#include <libcouchbase/couchbase.h>
#include <libcouchbase/api3.h>
#include <assert.h>
#include <string.h>
#include <vector>

static void
op_callback(lcb_t, int cbtype, const lcb_RESPBASE *rb)
{
    fprintf(stderr, "Got callback for %s.. ", lcb_strcbtype(cbtype));
    if (rb->rc != LCB_SUCCESS && rb->rc != LCB_SUBDOC_MULTI_FAILURE) {
        fprintf(stderr, "Operation failed rc: %x (%s)\n", rb->rc, lcb_strerror(NULL, rb->rc));
        return;
    }

    if (cbtype == LCB_CALLBACK_GET) {
        const lcb_RESPGET *rg = reinterpret_cast<const lcb_RESPGET*>(rb);
        fprintf(stderr, "Value %.*s\n", (int)rg->nvalue, rg->value);
    } else if (cbtype == LCB_CALLBACK_SDMUTATE || cbtype == LCB_CALLBACK_SDLOOKUP) {
        const lcb_RESPSUBDOC *resp = reinterpret_cast<const lcb_RESPSUBDOC*>(rb);
        lcb_SDENTRY ent;
        size_t iter = 0;
        if (lcb_sdresult_next(resp, &ent, &iter)) {
            fprintf(stderr, "Status: 0x%x. Value: %.*s\n", ent.status, (int)ent.nvalue, ent.value);
        } else {
            fprintf(stderr, "No result!\n");
        }
    } else {
        fprintf(stderr, "OK\n");
    }
}

// Function to issue an lcb_get3() (and print the state of the document)
static void
demoKey(lcb_t instance, const char *key)
{
    printf("Retrieving '%s'\n", key);
    printf("====\n");
    lcb_CMDGET gcmd = { 0 };
    LCB_CMD_SET_KEY(&gcmd, key, strlen(key));
    lcb_error_t rc = lcb_get3(instance, NULL, &gcmd);
    assert(rc == LCB_SUCCESS);
    lcb_wait(instance);
    printf("====\n\n");
}

// cluster_run mode
#define DEFAULT_CONNSTR "couchbase://localhost"
int main(int argc, char **argv)
{
    lcb_create_st crst = { 0 };
    crst.version = 3;
    if (argc > 1) {
        crst.v.v3.connstr = argv[1];
    } else {
        crst.v.v3.connstr = DEFAULT_CONNSTR;
    }

    std::string pass("asdasd");

    crst.v.v3.type = LCB_TYPE_BUCKET;
    crst.v.v3.passwd = pass.c_str();

    lcb_t instance;
    lcb_error_t rc = lcb_create(&instance, &crst);
    assert(rc == LCB_SUCCESS);

    rc = lcb_connect(instance);
    assert(rc == LCB_SUCCESS);

    lcb_wait(instance);

    rc = lcb_get_bootstrap_status(instance);
    assert(rc == LCB_SUCCESS);

    lcb_install_callback3(instance, LCB_CALLBACK_DEFAULT, op_callback);

    // Store the initial document. Subdocument operations cannot create
    // documents
    printf("Storing the initial item..\n");
    // Store an item
    lcb_CMDSTORE scmd = { 0 };
    scmd.operation = LCB_SET;
    LCB_CMD_SET_KEY(&scmd, "ram", 3);
    const char *initval = "{\"hello\":\"world\"}";
    LCB_CMD_SET_VALUE(&scmd, initval, strlen(initval));
    rc = lcb_store3(instance, NULL, &scmd);
    assert(rc == LCB_SUCCESS);
    lcb_wait(instance);

    // xattr change
    lcb_CMDSUBDOC mcmd = {0};
    LCB_CMD_SET_KEY(&mcmd, "ram", 3);

    std::vector<lcb_SDSPEC> specs;

    lcb_SDSPEC spec1, spec2 = {0};
    spec1.sdcmd = LCB_SDCMD_DICT_UPSERT;
    spec1.options = LCB_SDSPEC_F_MKINTERMEDIATES | LCB_SDSPEC_F_XATTR_MACROVALUES;
    LCB_SDSPEC_SET_PATH(&spec1, "_eventing.cas", 13);
    static const std::string MUTATION_CAS_MACRO("\"${Mutation.CAS}\"");
    LCB_SDSPEC_SET_VALUE(&spec1, MUTATION_CAS_MACRO.c_str(), MUTATION_CAS_MACRO.size());
    specs.push_back(spec1);

    spec1.sdcmd = LCB_SDCMD_ARRAY_ADD_LAST;
    spec1.options = LCB_SDSPEC_F_MKINTERMEDIATES | LCB_SDSPEC_F_XATTRPATH ;
    LCB_SDSPEC_SET_PATH(&spec1, "_eventing.timers", 16);
    LCB_SDSPEC_SET_VALUE(&spec1, "\"12:00:00\"", 10);
    specs.push_back(spec1);

    spec2.sdcmd = LCB_SDCMD_SET_FULLDOC;
    LCB_SDSPEC_SET_PATH(&spec2, "", 0);
    LCB_SDSPEC_SET_VALUE(&spec2, "{\"val\":\"1234\"}", 14);
    specs.push_back(spec2);

    mcmd.specs = specs.data();
    mcmd.nspecs = specs.size();
    mcmd.cmdflags = LCB_CMDSUBDOC_F_UPSERT_DOC;

    rc = lcb_subdoc3(instance, NULL, &mcmd);
    assert(rc == LCB_SUCCESS);
    lcb_wait(instance);
    demoKey(instance, "ram");

    lcb_destroy(instance);
    return 0;
}
