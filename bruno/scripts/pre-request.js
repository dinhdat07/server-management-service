// Auto-inject Bearer token for all requests
const token = bru.getVar("accessToken");
if (token) {
  req.setHeader("Authorization", "Bearer " + token);
}

// Auto-inject CSRF token for mutation requests
const csrfToken = bru.getVar("csrfToken");
if (csrfToken) {
  req.setHeader("X-CSRF-Token", csrfToken);
}

console.log("[pre-request] " + req.getMethod() + " " + req.getUrl());
