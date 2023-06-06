STRIPE_PUBLIC_KEY='pk_test_51NAAa2H05TTfxLt3LbCIUakd1KxUirl71AdAJ7yg0xXuoxYVzwIZYeaiLiNgubGu4EBEElyAaemJdCY5S24TJ3C700jIf4fsfg'
STRIPE_SECRET_KEY='HERE_SHOULD_BE_STRIPE_SECRET_KEY'
FRONTEND_PORT=4000
BACKEND_PORT=4001
DSN=root@tcp(localhost:3306)/widgets?parseTime=true&tls=false


build: clean build_frontend build_backend
	@printf "All binaries built!\n"


clean:
	@- rm -f dist/*
	@go clean


build_frontend:
	@echo "building frontend..."
	@go build -o dist/purchases ./cmd/web
	@echo "frontend was build"


build_backend:
	@echo "building backend..."
	@go build -o dist/purchases_api ./cmd/api
	@echo "backend was build"


start_frontend: build_frontend
	@echo "starting frontend..."
	@env STRIPE_PUBLIC_KEY=${STRIPE_PUBLIC_KEY} STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY} ./dist/purchases -port=${FRONTEND_PORT} &
	@echo "frontend running"


start_backend: build_backend
	@echo "starting backend..."
	@env STRIPE_PUBLIC_KEY=${STRIPE_PUBLIC_KEY} STRIPE_SECRET_KEY=${STRIPE_SECRET_KEY} ./dist/purchases_api -port=${BACKEND_PORT} &
	@echo "backend running"


start: start_frontend start_backend


stop_frontend:
	@-pkill -SIGTERM -f "purchases -port=${FRONTEND_PORT}"


stop_backend:
	@-pkill -SIGTERM -f "purchases_api -port=${BACKEND_PORT}"


stop: stop_frontend stop_backend