function queryParams() {
  var username = localStorage.getItem("username");
  var token = localStorage.getItem("token");
  return 'username=' + username + '&token=' + token;
}

function queryPath() {
  // var username = localStorage.getItem("username");
  // var token = localStorage.getItem("token");
  // var parentID = localStorage.getItem("parentID");
  // var documentName = localStorage.getItem("documentName");
  // return 'username=' + username + '&token=' + token + '&parentID=' + parentID + '&documentName=' + documentName ;

  return 'username=' + localStorage.getItem("username")
      + '&token=' + localStorage.getItem("token")
      + '&parentID=' + localStorage.getItem("parentID")
      + '&documentName=' + localStorage.getItem("documentName") ;
}