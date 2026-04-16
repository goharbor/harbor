# Multi-stage build for harbor-cleanup tool
FROM c.rzp.io/proxy_dockerhub/library/python:3.10-alpine as builder

# Set working directory
WORKDIR /app

# Install system dependencies for building Python packages
RUN apk add --no-cache \
    gcc \
    musl-dev \
    libffi-dev

# Copy requirements first for better Docker layer caching
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir --user -r requirements.txt

# Final stage - minimal runtime image
FROM c.rzp.io/proxy_dockerhub/library/python:3.10-alpine

# Set working directory
WORKDIR /app

# Create non-root user for security
RUN addgroup -g 1001 harborcleaner && \
    adduser -D -u 1001 -G harborcleaner harborcleaner

# Copy Python dependencies from builder stage
COPY --from=builder /root/.local /home/harborcleaner/.local

# Copy application code
COPY main.py .
COPY harbor_cleanup/ ./harbor_cleanup/

# Set proper permissions
RUN chown -R harborcleaner:harborcleaner /app

# Update PATH to include user-installed packages
ENV PATH=/home/harborcleaner/.local/bin:$PATH

# Switch to non-root user
USER harborcleaner

# Set default environment variables for container runs
ENV PYTHONUNBUFFERED=1
ENV LOG_LEVEL=INFO
ENV DRY_RUN=true
ENV VERBOSE_ARTIFACTS=false


# Default command
ENTRYPOINT ["python3", "main.py"] 