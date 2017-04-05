function OnUpdate(doc, meta) {

  log("doc: ", doc, " meta: ", meta);

  if (doc.type == "credit_score") {

      updated_doc = CalculateCreditScore(doc);
      log("updated doc: ", updated_doc);

      credit_bucket[meta.docid] = updated_doc;
      log("Wrote to credit_bucket, doc id: ", meta.docid);

      var value = credit_bucket[meta.docid];
      log("Fetched from credit_bucket, docid: ", meta.docid, " value:", value);

      //delete credit_bucket[meta.docid];


  } else {

      /*
      var bucket = "beer-sample";
      var limit = 5;
      var type = "brewery";

      var n1qlResult = n1ql("select ${bucket}.name from ${bucket} where ${bucket}.type == '${type}' limit ${limit}");
      var n1qlResultLength = n1qlResult.length;

      for (i = 0; i < n1qlResultLength; i++) {
          log("n1ql query response row: ", n1qlResult[i]);
      }*/
  }
}

function OnDelete(msg) {

}

function CalculateCreditScore(doc) {
  var credit_score = 500;

  if (doc.credit_limit_used/doc.total_credit_limit < 0.3) {
    credit_score = credit_score + 50;
  } else {
    doc.credit_score = doc.credit_score - Math.floor((doc.credit_limit_used/doc.total_credit_limit) * 20);
  }

  if (doc.missed_emi_payments !== 0) {
    credit_score = credit_score - doc.missed_emi_payments * 30;
  }

  if (credit_score < 300) {
    doc.credit_score = 300;
  } else {
    doc.credit_score = credit_score;
  }

  return doc;
}
