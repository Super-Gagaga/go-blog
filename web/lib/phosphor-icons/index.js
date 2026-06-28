// Phosphor Icons v2.0.3 — local copy (see web/lib/phosphor-icons/)
var head = document.getElementsByTagName("head")[0];

for (weight of ["regular", "thin", "light", "bold", "fill", "duotone"]) {
  var link = document.createElement("link");
  link.rel = "stylesheet";
  link.type = "text/css";
  link.href = "/lib/phosphor-icons/" + weight + "/style.css";
  head.appendChild(link);
}
