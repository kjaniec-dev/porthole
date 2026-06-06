import React, { useEffect, useState } from "https://esm.sh/react@18";
import { createRoot } from "https://esm.sh/react-dom@18/client";

const h = React.createElement;
const apiUrl = "/api/hello";
const directApiUrl = "https://api.localhost/hello";

function InfoCard(props) {
  return h(
    "section",
    { className: "card" },
    h("p", { className: "eyebrow" }, props.label),
    h("div", { className: "value" }, props.value),
    props.note ? h("p", { className: "note" }, props.note) : null
  );
}

function App() {
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState(null);
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      const response = await fetch(apiUrl);
      if (!response.ok) {
        throw new Error(`Backend returned ${response.status}`);
      }
      const body = await response.json();
      setData(body);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  return h(
    "main",
    { className: "app-shell" },
    h(
      "header",
      { className: "hero" },
      h("p", { className: "kicker" }, "Porthole Local Stack"),
      h("h1", null, "React frontend routed through Traefik"),
      h(
        "p",
        { className: "lede" },
        "This page is served by the example frontend container. Its data comes from the Python backend through a separate Traefik router."
      ),
      h(
        "div",
        { className: "actions" },
        h(
          "a",
          { href: apiUrl, target: "_blank", rel: "noreferrer" },
          "Open backend JSON"
        ),
        h("button", { type: "button", onClick: load }, "Reload data")
      )
    ),
    h(
      "section",
      { className: "grid" },
      h(InfoCard, {
        label: "Frontend URL",
        value: "https://frontend.localhost",
        note: "Served through Traefik over TLS",
      }),
      h(InfoCard, {
        label: "Backend URL",
        value: directApiUrl,
        note: "Also available through the same-origin /api router",
      }),
      h(InfoCard, {
        label: "State",
        value: loading ? "Loading..." : error ? "Error" : "Connected",
        note: error || "Backend data fetched successfully",
      })
    ),
    h(
      "section",
      { className: "card response" },
      h("p", { className: "eyebrow" }, "Backend response"),
      h(
        "pre",
        null,
        JSON.stringify(
          data || { message: "Waiting for backend response..." },
          null,
          2
        )
      )
    )
  );
}

createRoot(document.getElementById("root")).render(h(App));
