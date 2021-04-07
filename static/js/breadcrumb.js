

var path = [
    "Home",
]

function storageObj(obj) {
    var checkedIdStr = JSON.stringify(obj);
    localStorage.setItem("Path", checkedIdStr);
};


// storageObj(path);
function initStorageObj() {
    storageObj(path);
}
// JSON.parse( localStorage.getItem("Path") );
// localStorage.setItem( 'Path' , path )


