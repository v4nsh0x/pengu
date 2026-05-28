document.addEventListener("DOMContentLoaded", () => {
    // Hero window reveal animation
    setTimeout(() => {
        const codeWindow = document.querySelector('.code-window');
        if (codeWindow) {
            codeWindow.classList.add('visible');
        }
    }, 300);

    // Section scroll animations
    const sections = document.querySelectorAll('section');
    const navLinks = document.querySelectorAll('.sidebar a');

    // Observer for fading in sections as you scroll
    const observerOptions = {
        root: null,
        rootMargin: '0px',
        threshold: 0.15
    };

    const sectionObserver = new IntersectionObserver((entries, observer) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('visible');
            }
        });
    }, observerOptions);

    sections.forEach(section => {
        sectionObserver.observe(section);
    });

    // Observer for highlighting active sidebar link
    const highlightOptions = {
        root: null,
        rootMargin: '-20% 0px -70% 0px',
        threshold: 0
    };

    const highlightObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const id = entry.target.getAttribute('id');
                navLinks.forEach(link => {
                    link.classList.remove('active');
                    if (link.getAttribute('href') === `#${id}`) {
                        link.classList.add('active');
                    }
                });
            }
        });
    }, highlightOptions);

    sections.forEach(section => {
        highlightObserver.observe(section);
    });

    // Smooth scroll for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            const targetId = this.getAttribute('href');
            if (targetId === '#') return;
            
            const targetElement = document.querySelector(targetId);
            if (targetElement) {
                e.preventDefault();
                targetElement.scrollIntoView({
                    behavior: 'smooth'
                });
                
                // Update URL without jump
                history.pushState(null, null, targetId);
            }
        });
    });

    // Install tab switching
    const installTabs = document.querySelectorAll('.install-tab');
    const installPanels = document.querySelectorAll('.install-panel');

    installTabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const target = tab.getAttribute('data-tab');

            installTabs.forEach(t => t.classList.remove('active'));
            installPanels.forEach(p => p.classList.remove('active'));

            tab.classList.add('active');
            document.getElementById(`panel-${target}`).classList.add('active');
        });
    });
});
