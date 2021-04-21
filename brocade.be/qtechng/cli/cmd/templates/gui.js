
function syntaxHighlight(json) {
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        var cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}


function handleCheck(id) {
    var h = document.getElementById(id);
    var v0 = h.getAttribute("data-off");
    var v1 = h.getAttribute("data-on");
    var v = h.value;
    h.value = v == v1 ? v0 : v1;
    return true
}

function toggle(id) {
    var elm = document.getElementById(id);
    if (elm.classList.contains("visible")) {
        elm.classList.remove("visible");
        elm.classList.add("invisible");
    } else {
        elm.classList.remove("invisible");
        elm.classList.add("visible");
    }
}
