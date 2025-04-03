document.addEventListener('DOMContentLoaded', function() {
    // Get all copy buttons
    const copyButtons = document.querySelectorAll('.copy-btn');
    
    // Add click event listeners to each button
    copyButtons.forEach(button => {
        button.addEventListener('click', function() {
            // Get the URL to copy
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
            
            // Change button text to 'Copied!' briefly
            const originalText = this.textContent;
            this.textContent = 'Copied!';
            this.style.backgroundColor = '#28a745';
            
            // Reset button text after 2 seconds
            setTimeout(() => {
                this.textContent = originalText;
                this.style.backgroundColor = '';
            }, 2000);
        });
    });
});