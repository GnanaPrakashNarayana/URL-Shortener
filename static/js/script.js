document.addEventListener('DOMContentLoaded', function() {

    // --- Ripple Effect ---
    const buttons = document.querySelectorAll('.btn, .copy-btn, .oauth-btn');
    buttons.forEach(button => {
        button.addEventListener('click', function(e) {
            // Clear existing ripples first
            const existingRipple = button.querySelector('.ripple');
            if (existingRipple) {
                existingRipple.remove();
            }

            const rect = e.target.getBoundingClientRect();
            const x = e.clientX - rect.left;
            const y = e.clientY - rect.top;

            const ripple = document.createElement('span');
            ripple.classList.add('ripple');
            ripple.style.left = `${x}px`;
            ripple.style.top = `${y}px`;

            // Use requestAnimationFrame for smoother animation
            requestAnimationFrame(() => {
                this.appendChild(ripple);
            });

            // Clean up ripple after animation
            setTimeout(() => {
                if (ripple.parentNode) { 
                   ripple.remove();
                }
            }, 600);
        });
    });

    // --- URL Expansion Feature ---
    const setupUrlExpansion = () => {
        const urlLinks = document.querySelectorAll('.url-link');
        
        urlLinks.forEach(link => {
            // Remove any existing indicators first
            const existingIndicator = link.querySelector('.url-expand-indicator');
            if (existingIndicator) {
                existingIndicator.remove();
            }
            
            // Check if the content is truncated
            const isTruncated = link.scrollWidth > link.clientWidth;
            
            // Add expand indicator if truncated
            if (isTruncated) {
                const indicator = document.createElement('span');
                indicator.className = 'url-expand-indicator';
                indicator.textContent = '···';
                link.appendChild(indicator);
                
                // Add truncated class for styling
                link.classList.add('truncated');
                
                // Add click handler for the indicator
                indicator.addEventListener('click', function(e) {
                    e.preventDefault();
                    e.stopPropagation();
                    showUrlPopup(link);
                });
                
                // Add right-click (contextmenu) handler to show full URL
                link.addEventListener('contextmenu', function(e) {
                    // Don't prevent default context menu, but show popup
                    showUrlPopup(link);
                });
                
                // Add alt+click handler to show URL without navigating
                link.addEventListener('click', function(e) {
                    // If Alt key is pressed, show the popup instead of navigating
                    if (e.altKey) {
                        e.preventDefault();
                        showUrlPopup(link);
                    }
                    // Otherwise, let the default link behavior happen
                });
            } else {
                link.classList.remove('truncated');
            }
        });
    };
    
    // Function to show URL popup
    function showUrlPopup(linkElement) {
        // Remove any existing popup
        document.querySelectorAll('.url-popup').forEach(p => p.remove());

        // Get the bounding rect of the clicked element
        const rect = linkElement.getBoundingClientRect();

        // Create the popup
        const popup = document.createElement('div');
        popup.className = 'url-popup';
        popup.textContent = linkElement.getAttribute('data-url') || linkElement.getAttribute('href') || linkElement.textContent;

        // Position the popup
        let top = rect.bottom + 5;
        let left = rect.left;
        
        // Adjust position if too close to bottom of screen
        if (window.innerHeight - rect.bottom < 120) {
            top = rect.top - 80; // Position above the link
        }
        
        // Adjust position if too close to right edge of screen
        if (window.innerWidth - rect.left < 350) {
            left = Math.max(10, window.innerWidth - 460);
        }
        
        popup.style.top = `${top}px`;
        popup.style.left = `${left}px`;

        document.body.appendChild(popup);

        // Remove popup on outside click
        const outsideClickHandler = (event) => {
            if (!popup.contains(event.target)) {
                popup.remove();
                document.removeEventListener('click', outsideClickHandler);
            }
        };
        
        setTimeout(() => {
            document.addEventListener('click', outsideClickHandler);
        }, 10);
    }
    
    // On window resize, remove any popups and recalculate truncation
    window.addEventListener('resize', () => {
        document.querySelectorAll('.url-popup').forEach(p => p.remove());
        setTimeout(setupUrlExpansion, 300); // Re-check truncation after resize
    });
    
    // Setup URL expansion initially
    setTimeout(setupUrlExpansion, 500);
    
    // Re-check periodically
    setInterval(setupUrlExpansion, 3000);

    // --- Form Validation & Animations ---
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        const requiredInputs = form.querySelectorAll('[required]');

        // Function to show error message
        const showError = (input, message) => {
            input.classList.add('invalid');
            const formGroup = input.closest('.form-group');
            if (formGroup) {
                // Remove existing error message first
                const existingError = formGroup.querySelector('.error-msg');
                if (existingError) {
                    existingError.remove();
                }
                // Add new error message
                const errorMsg = document.createElement('div');
                errorMsg.className = 'error-msg';
                errorMsg.textContent = message || 'This field is required';
                formGroup.appendChild(errorMsg);

                // Add shake animation to the input
                input.classList.add('shake');
                // Remove shake class after animation finishes
                input.addEventListener('animationend', () => {
                    input.classList.remove('shake');
                }, { once: true });
            }
        };

        // Function to clear error message
        const clearError = (input) => {
            input.classList.remove('invalid');
            input.classList.remove('shake');
            const formGroup = input.closest('.form-group');
            if (formGroup) {
                const errorMsg = formGroup.querySelector('.error-msg');
                if (errorMsg) {
                    errorMsg.remove();
                }
            }
        };

        // Validate on Submit
        form.addEventListener('submit', function(event) {
            let isValid = true;
            requiredInputs.forEach(input => {
                clearError(input); // Clear previous errors
                if (!input.value.trim()) {
                    isValid = false;
                    showError(input);
                }
                // Add specific validation like email format if needed
                if (input.type === 'email' && input.value.trim() && !/\S+@\S+\.\S+/.test(input.value)) {
                     isValid = false;
                     showError(input, 'Please enter a valid email address.');
                }
                // Add password confirmation check
                if (input.id === 'password_confirm') {
                    const passwordInput = form.querySelector('#password');
                    if (passwordInput && input.value !== passwordInput.value) {
                         isValid = false;
                         showError(input, 'Passwords do not match.');
                    }
                }
            });

            if (!isValid) {
                event.preventDefault(); // Prevent form submission
            }
        });

        // Real-time validation feedback (on input change)
        requiredInputs.forEach(input => {
            input.addEventListener('input', function() {
                // Clear error as user types
                if (this.classList.contains('invalid')) {
                    clearError(this);
                }
            });
            
            input.addEventListener('blur', function() {
                // Optionally re-validate on blur
                clearError(this); // Clear first
                
                if (this.hasAttribute('required') && !this.value.trim()) {
                    showError(this);
                } else if (this.type === 'email' && this.value.trim() && !/\S+@\S+\.\S+/.test(this.value)) {
                     showError(this, 'Please enter a valid email address.');
                } else if (this.id === 'password_confirm') {
                    const passwordInput = form.querySelector('#password');
                    if (passwordInput && this.value && this.value !== passwordInput.value) {
                         showError(this, 'Passwords do not match.');
                    }
                }
            });
        });
    });

    // --- Copy Button Functionality ---
    const copyButtons = document.querySelectorAll('.copy-btn');
    copyButtons.forEach(button => {
        button.addEventListener('click', function(e) {
            e.stopPropagation(); // Prevent event bubbling
            
            const url = this.getAttribute('data-url');
            if (!url) return;

            navigator.clipboard.writeText(url).then(() => {
                // Success feedback
                const originalText = this.textContent;
                this.textContent = 'Copied!';
                this.style.backgroundColor = 'var(--success-color)';
                this.style.color = 'white';
                this.classList.add('copied');

                // Reset button after a delay
                setTimeout(() => {
                    this.textContent = originalText;
                    this.style.backgroundColor = '';
                    this.style.color = '';
                    this.classList.remove('copied');
                }, 1500);
            }).catch(err => {
                console.error('Failed to copy URL: ', err);
            });
        });
    });

    // --- Advanced Options Toggle - IMPROVED ---
    const advancedOptionsToggles = document.querySelectorAll('.advanced-options-toggle');
    
    advancedOptionsToggles.forEach(toggle => {
        toggle.addEventListener('click', function(e) {
            e.preventDefault(); 
            e.stopPropagation();
            
            // Find the parent .advanced-options element
            const parent = this.closest('.advanced-options');
            if (!parent) return;
            
            // Find the content container
            const content = parent.querySelector('.advanced-options-content');
            if (!content) return;
            
            // Toggle active class on both the button and content
            this.classList.toggle('active');
            
            if (this.classList.contains('active')) {
                // Open the advanced options with a smooth animation
                content.classList.add('active');
                // Update the icon
                const icon = this.querySelector('.toggle-icon');
                if (icon) {
                    icon.textContent = '−';
                }
                
                // Scroll to show the content if needed
                setTimeout(() => {
                    const rect = content.getBoundingClientRect();
                    const isVisible = (
                        rect.top >= 0 &&
                        rect.bottom <= window.innerHeight
                    );
                    
                    if (!isVisible) {
                        content.scrollIntoView({
                            behavior: 'smooth',
                            block: 'nearest'
                        });
                    }
                }, 300); // Wait for animation to start
            } else {
                // Close the advanced options
                content.classList.remove('active');
                // Update the icon
                const icon = this.querySelector('.toggle-icon');
                if (icon) {
                    icon.textContent = '+';
                }
            }
        });
    });

    // Initial check for animations on page elements
    const animateOnScroll = () => {
        const cards = document.querySelectorAll('.card:not(.auth-card), .stat-card');
        const screenPosition = window.innerHeight / 1.2;

        cards.forEach(card => {
            const cardPosition = card.getBoundingClientRect().top;
            if (!card.classList.contains('animated') && cardPosition < screenPosition) {
                card.classList.add('animated', 'animate-fadeInUp');
            }
        });
    };

    // Run animation check on page load and scroll
    if (document.querySelector('.card:not(.auth-card), .stat-card')) {
        animateOnScroll();
        window.addEventListener('scroll', animateOnScroll);
    }

    // Make tables responsive
    const tables = document.querySelectorAll('.urls-table');
    if (tables.length > 0) {
        // Add horizontal scrolling container if needed
        tables.forEach(table => {
            const wrapper = table.parentElement;
            if (!wrapper.classList.contains('table-responsive') && window.innerWidth <= 768) {
                wrapper.style.overflowX = 'auto';
                wrapper.style.WebkitOverflowScrolling = 'touch';
                wrapper.style.marginBottom = '1rem';
            }
        });
    }
});

// Password Toggle Functionality
document.addEventListener('DOMContentLoaded', function() {
    const passwordInputs = document.querySelectorAll('input[type="password"]');
    
    passwordInputs.forEach(input => {
        // Create toggle button if not already present
        if (!input.parentNode.classList.contains('password-input-group')) {
            // Create wrapper for the input
            const wrapper = document.createElement('div');
            wrapper.className = 'password-input-group';
            
            // Insert the wrapper before the input
            input.parentNode.insertBefore(wrapper, input);
            
            // Move the input into the wrapper
            wrapper.appendChild(input);
            
            // Create toggle button
            const toggleBtn = document.createElement('button');
            toggleBtn.type = 'button';
            toggleBtn.className = 'password-toggle';
            toggleBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>';
            
            // Add the toggle button to the wrapper
            wrapper.appendChild(toggleBtn);
            
            // Add click event to toggle password visibility
            toggleBtn.addEventListener('click', function(e) {
                e.preventDefault();
                const type = input.getAttribute('type') === 'password' ? 'text' : 'password';
                input.setAttribute('type', type);
                
                // Change the icon based on visibility
                if (type === 'text') {
                    this.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path><line x1="1" y1="1" x2="23" y2="23"></line></svg>';
                } else {
                    this.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>';
                }
            });
        }
    });

    // Calculate totals for dashboard stats when needed
    if (document.getElementById('totalVisits')) {
        updateStats();
    }

    function updateStats() {
        const visitCounts = document.querySelectorAll('.visit-count');
        let totalVisits = 0;
        
        visitCounts.forEach(function(element) {
            totalVisits += parseInt(element.getAttribute('data-visits') || 0);
        });
        
        const totalVisitsElement = document.getElementById('totalVisits');
        if (totalVisitsElement) {
            totalVisitsElement.textContent = totalVisits;
        }
        
        // Calculate active links
        const expiryDates = document.querySelectorAll('.expiry-date');
        let activeLinks = 0;
        
        expiryDates.forEach(function(element) {
            if (element.getAttribute('data-expired') !== 'true') {
                activeLinks++;
            }
        });
        
        const activeLinksElement = document.getElementById('activeLinks');
        if (activeLinksElement) {
            activeLinksElement.textContent = activeLinks;
        }
    }
});

// QR Code Toggle and Preview Functionality
document.addEventListener('DOMContentLoaded', function() {
    const qrToggle = document.getElementById('qr-code-toggle');
    const qrPreview = document.querySelector('.qr-code-preview');
    const urlInput = document.getElementById('url-input');
    
    if (qrToggle && qrPreview) {
        qrToggle.addEventListener('change', function() {
            if (this.checked) {
                qrPreview.classList.add('active');
                
                // Only update preview if URL is valid
                if (urlInput && urlInput.value && isValidURL(urlInput.value)) {
                    updateQRPreview(urlInput.value);
                }
            } else {
                qrPreview.classList.remove('active');
            }
        });
    }
    
    // Update QR preview when URL changes
    if (urlInput && qrPreview) {
        urlInput.addEventListener('input', function() {
            if (qrToggle && qrToggle.checked && isValidURL(this.value)) {
                updateQRPreview(this.value);
            }
        });
        
        // Debounce function for performance
        let typingTimer;
        urlInput.addEventListener('keyup', function() {
            clearTimeout(typingTimer);
            if (qrToggle && qrToggle.checked && this.value) {
                typingTimer = setTimeout(function() {
                    if (isValidURL(urlInput.value)) {
                        updateQRPreview(urlInput.value);
                    }
                }, 500);
            }
        });
    }
    
    // Format change handler
    const formatOptions = document.querySelectorAll('input[name="qr_format"]');
    formatOptions.forEach(option => {
        option.addEventListener('change', function() {
            if (qrToggle && qrToggle.checked && urlInput && isValidURL(urlInput.value)) {
                updateQRPreview(urlInput.value);
            }
        });
    });
    
    // QR Code buttons in tables
    const qrButtons = document.querySelectorAll('.qr-code-btn');
    qrButtons.forEach(button => {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            e.stopPropagation();
            
            const url = this.getAttribute('data-url');
            if (url) {
                const id = url.split('/').pop();
                window.location.href = `/qrcode/preview/${id}?base_url=${encodeURIComponent(url.replace('/' + id, ''))}`;
            }
        });
    });
});

// Function to check if URL is valid
function isValidURL(string) {
    try {
        new URL(string);
        return true;
    } catch (_) {
        return false;
    }
}

// Function to update QR preview
function updateQRPreview(url) {
    const qrPreviewImg = document.querySelector('.qr-preview-image');
    if (!qrPreviewImg) return;
    
    // Get selected format
    let format = 'png';
    const formatOptions = document.querySelectorAll('input[name="qr_format"]');
    formatOptions.forEach(option => {
        if (option.checked) {
            format = option.value;
        }
    });
    
    // Set loading state
    qrPreviewImg.style.opacity = '0.5';
    
    // Create a temporary URL using the API
    const apiUrl = `/qrcode/generate?url=${encodeURIComponent(url)}&format=${format}`;
    
    // Update preview with slight delay for animation effect
    setTimeout(() => {
        qrPreviewImg.src = apiUrl;
        qrPreviewImg.style.opacity = '1';
    }, 200);
}