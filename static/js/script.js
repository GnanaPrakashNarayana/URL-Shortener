document.addEventListener('DOMContentLoaded', function() {

    // --- Ripple Effect ---
    // (Kept from original - CSS handles the visuals)
    const buttons = document.querySelectorAll('.btn, .copy-btn, .oauth-btn'); // Added oauth-btn
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

            // Use requestAnimationFrame to ensure the element is added before animation starts
            requestAnimationFrame(() => {
                this.appendChild(ripple);
            });


            // Clean up ripple after animation
            setTimeout(() => {
                if (ripple.parentNode) { // Check if still attached
                   ripple.remove();
                }
            }, 600);
        });
    });


    // --- Form Validation & Animations ---
    const forms = document.querySelectorAll('form.auth-form'); // Target auth forms specifically if needed
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
                errorMsg.textContent = message || 'This field is required'; // Default message
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
             input.classList.remove('shake'); // Ensure shake is removed
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


     // --- Copy Button Functionality (for home/dashboard pages) ---
    const copyButtons = document.querySelectorAll('.copy-btn');
    copyButtons.forEach(button => {
        button.addEventListener('click', function() {
            const url = this.getAttribute('data-url');
            if (!url) return;

            navigator.clipboard.writeText(url).then(() => {
                // Success feedback
                const originalText = this.textContent;
                this.textContent = 'Copied!';
                this.style.backgroundColor = 'var(--success-color)'; // Use CSS variable
                this.style.color = 'white';

                 // Optional: Add a temporary success class
                 this.classList.add('copied');

                // Reset button after a delay
                setTimeout(() => {
                    this.textContent = originalText;
                    this.style.backgroundColor = ''; // Revert to default
                    this.style.color = ''; // Revert to default
                     this.classList.remove('copied');
                }, 1500); // Shorter duration
            }).catch(err => {
                console.error('Failed to copy URL: ', err);
                // Optional: Show error feedback to the user
            });
        });
    });


    // --- Remove fade-in and other general page animations from JS ---
    // CSS now handles the entry animations for auth pages.
    // If you need JS-driven animations for other pages (like scroll-triggered), keep that logic.
    // For example, the original script had IntersectionObserver logic.
    // If you want animations on the home/dashboard pages, keep or adapt that section.
    // Example: Keep scroll animations for cards on other pages
    const animateOnScroll = () => {
        const cards = document.querySelectorAll('.card:not(.auth-card), .stat-card'); // Exclude auth-card
        const screenPosition = window.innerHeight / 1.2; // Adjust trigger point

        cards.forEach(card => {
            const cardPosition = card.getBoundingClientRect().top;
            if (!card.classList.contains('animated') && cardPosition < screenPosition) {
                card.classList.add('animated', 'animate-fadeInUp'); // Use fadeInUp or scaleIn
            }
        });
    };

    // Initial check and add scroll listener if non-auth cards exist
     if (document.querySelector('.card:not(.auth-card), .stat-card')) {
        animateOnScroll(); // Run on load
        window.addEventListener('scroll', animateOnScroll);
    }


}); // End DOMContentLoaded