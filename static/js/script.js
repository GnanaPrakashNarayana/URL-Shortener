document.addEventListener('DOMContentLoaded', function() {
    // Apply fade-in animations to elements
    const fadeElements = document.querySelectorAll('.fade-in');
    fadeElements.forEach(element => {
        element.style.opacity = '0';
        
        // Create an observer for each element
        const observer = new IntersectionObserver(entries => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    // Add the animation class when element is in viewport
                    entry.target.classList.add('animate');
                    observer.unobserve(entry.target);
                }
            });
        }, { threshold: 0.1 });
        
        observer.observe(element);
    });
    
    // URL input animation
    const urlInput = document.querySelector('.url-input');
    if (urlInput) {
        urlInput.addEventListener('focus', function() {
            this.parentElement.classList.add('input-focused');
        });
        
        urlInput.addEventListener('blur', function() {
            this.parentElement.classList.remove('input-focused');
        });
    }
    
    // Copy button functionality with animation
    const copyButtons = document.querySelectorAll('.copy-btn');
    copyButtons.forEach(button => {
        button.addEventListener('click', function() {
            const url = this.getAttribute('data-url');
            
            // Create a temporary input element
            const input = document.createElement('input');
            input.value = url;
            document.body.appendChild(input);
            
            // Select the text
            input.select();
            input.setSelectionRange(0, 99999);
            
            // Copy to clipboard
            document.execCommand('copy');
            
            // Remove the temporary input
            document.body.removeChild(input);
            
            // Animate the button
            const originalText = this.textContent;
            this.textContent = 'Copied!';
            this.style.backgroundColor = 'var(--success-color)';
            this.style.color = 'white';
            
            // Add ripple effect
            const ripple = document.createElement('span');
            ripple.classList.add('ripple');
            this.appendChild(ripple);
            
            setTimeout(() => {
                ripple.remove();
            }, 600);
            
            // Reset button after animation
            setTimeout(() => {
                this.textContent = originalText;
                this.style.backgroundColor = '';
                this.style.color = '';
            }, 2000);
        });
    });
    
    // Form validation with animations
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
        form.addEventListener('submit', function(event) {
            const requiredInputs = form.querySelectorAll('[required]');
            let valid = true;
            
            requiredInputs.forEach(input => {
                if (!input.value.trim()) {
                    valid = false;
                    input.classList.add('invalid');
                    
                    // Add shake animation
                    input.classList.add('shake');
                    setTimeout(() => {
                        input.classList.remove('shake');
                    }, 600);
                    
                    const formGroup = input.closest('.form-group');
                    if (formGroup) {
                        const errorMsg = formGroup.querySelector('.error-msg') || document.createElement('div');
                        errorMsg.className = 'error-msg';
                        errorMsg.textContent = 'This field is required';
                        
                        if (!formGroup.querySelector('.error-msg')) {
                            formGroup.appendChild(errorMsg);
                        }
                    }
                } else {
                    input.classList.remove('invalid');
                    const formGroup = input.closest('.form-group');
                    if (formGroup) {
                        const errorMsg = formGroup.querySelector('.error-msg');
                        if (errorMsg) {
                            errorMsg.remove();
                        }
                    }
                }
            });
            
            if (!valid) {
                event.preventDefault();
            }
        });
        
        // Real-time validation
        const inputs = form.querySelectorAll('input, textarea');
        inputs.forEach(input => {
            input.addEventListener('input', function() {
                if (this.hasAttribute('required') && !this.value.trim()) {
                    this.classList.add('invalid');
                } else {
                    this.classList.remove('invalid');
                    const formGroup = this.closest('.form-group');
                    if (formGroup) {
                        const errorMsg = formGroup.querySelector('.error-msg');
                        if (errorMsg) {
                            errorMsg.remove();
                        }
                    }
                }
            });
        });
    });
    
    // Add smooth scroll behavior
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function(e) {
            e.preventDefault();
            const targetId = this.getAttribute('href');
            const targetElement = document.querySelector(targetId);
            
            if (targetElement) {
                window.scrollTo({
                    top: targetElement.offsetTop - 80,
                    behavior: 'smooth'
                });
            }
        });
    });
    
    // Add animation to cards on scroll
    const animateOnScroll = () => {
        const cards = document.querySelectorAll('.card, .url-card, .stat-card');
        
        cards.forEach(card => {
            const cardPosition = card.getBoundingClientRect().top;
            const screenPosition = window.innerHeight / 1.3;
            
            if (cardPosition < screenPosition) {
                card.classList.add('animate');
            }
        });
    };
    
    // Call the animation function on load and scroll
    animateOnScroll();
    window.addEventListener('scroll', animateOnScroll);
    
    // Add ripple effect to buttons
    const buttons = document.querySelectorAll('.btn, .copy-btn');
    buttons.forEach(button => {
        button.addEventListener('click', function(e) {
            const x = e.clientX - e.target.getBoundingClientRect().left;
            const y = e.clientY - e.target.getBoundingClientRect().top;
            
            const ripple = document.createElement('span');
            ripple.classList.add('ripple');
            ripple.style.left = `${x}px`;
            ripple.style.top = `${y}px`;
            
            this.appendChild(ripple);
            
            setTimeout(() => {
                ripple.remove();
            }, 600);
        });
    });
    
    // Add CSS for the ripple effect
    const style = document.createElement('style');
    style.textContent = `
        .btn, .copy-btn {
            position: relative;
            overflow: hidden;
        }
        
        .ripple {
            position: absolute;
            background: rgba(255, 255, 255, 0.7);
            border-radius: 50%;
            transform: scale(0);
            animation: ripple 0.6s linear;
            pointer-events: none;
        }
        
        @keyframes ripple {
            to {
                transform: scale(4);
                opacity: 0;
            }
        }
        
        .animate {
            animation: fadeIn 0.5s ease-out forwards;
        }
        
        .shake {
            animation: shake 0.5s cubic-bezier(0.36, 0.07, 0.19, 0.97) both;
        }
        
        .invalid {
            border-color: var(--danger-color) !important;
        }
        
        .error-msg {
            color: var(--danger-color);
            font-size: 0.8rem;
            margin-top: 4px;
            animation: fadeIn 0.3s ease-out forwards;
        }
        
        .input-focused {
            transform: scale(1.01);
        }
    `;
    document.head.appendChild(style);
});