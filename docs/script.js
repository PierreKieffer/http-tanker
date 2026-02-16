(() => {
  "use strict";

  // ── Smooth scroll for anchor links ──
  document.querySelectorAll('a[href^="#"]').forEach((link) => {
    link.addEventListener("click", (e) => {
      const id = link.getAttribute("href").slice(1);
      const target = document.getElementById(id);
      if (target) {
        e.preventDefault();
        target.scrollIntoView({ behavior: "smooth" });
        // Close mobile nav if open
        document.querySelector(".navbar-links")?.classList.remove("open");
      }
    });
  });

  // ── Mobile nav toggle ──
  const toggle = document.querySelector(".nav-toggle");
  const navLinks = document.querySelector(".navbar-links");
  if (toggle && navLinks) {
    toggle.addEventListener("click", () => {
      navLinks.classList.toggle("open");
    });
  }

  // ── Active nav link on scroll ──
  const sections = document.querySelectorAll("section[id]");
  const navItems = document.querySelectorAll(".navbar-links a");

  function highlightNav() {
    const scrollY = window.scrollY + 100;
    let current = "";
    sections.forEach((section) => {
      if (section.offsetTop <= scrollY) {
        current = section.id;
      }
    });
    navItems.forEach((a) => {
      a.classList.toggle("active", a.getAttribute("href") === "#" + current);
    });
  }

  // ── Back to top ──
  const backToTop = document.querySelector(".back-to-top");

  function toggleBackToTop() {
    if (backToTop) {
      backToTop.classList.toggle("visible", window.scrollY > 300);
    }
  }

  if (backToTop) {
    backToTop.addEventListener("click", () => {
      window.scrollTo({ top: 0, behavior: "smooth" });
    });
  }

  window.addEventListener("scroll", () => {
    highlightNav();
    toggleBackToTop();
  }, { passive: true });

  highlightNav();

  // ── Copy to clipboard ──
  document.querySelectorAll(".copy-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      const block = btn.closest(".code-block");
      const code = block?.querySelector("pre")?.textContent || "";
      navigator.clipboard.writeText(code).then(() => {
        btn.classList.add("copied");
        const original = btn.innerHTML;
        btn.innerHTML = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg> Copied';
        setTimeout(() => {
          btn.classList.remove("copied");
          btn.innerHTML = original;
        }, 2000);
      });
    });
  });
})();
