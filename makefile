eth_service:
	@echo "Building eth_service"
	docker-compose -p eth_service \
	-f docker-compose.yml up -d --build 
	@echo "Done"
