package app

import (

	// loy:region:imports
	billingRepo "loymart/internal/billing/repository"
	billingService "loymart/internal/billing/service"
	billingHttp "loymart/internal/billing/transport/http"
	notificationRepo "loymart/internal/notification/repository"
	notificationService "loymart/internal/notification/service"
	notificationHttp "loymart/internal/notification/transport/http"
	organizationRepo "loymart/internal/organization/repository"
	organizationService "loymart/internal/organization/service"
	organizationHttp "loymart/internal/organization/transport/http"
	userRepo "loymart/internal/user/repository"
	userService "loymart/internal/user/service"
	userHttp "loymart/internal/user/transport/http"
	// loy:endregion
)

// wireDependencies initializes repositories, services, handlers and route bindings.
func (a *App) wireDependencies() error {
	// loy:region:repositories
	userRepo, _ := userRepo.NewPostgresRepository(a.db)
	_ = userRepo
	organizationRepo, _ := organizationRepo.NewPostgresRepository(a.db)
	_ = organizationRepo
	billingRepo, _ := billingRepo.NewPostgresRepository(a.db)
	_ = billingRepo
	notificationRepo, _ := notificationRepo.NewPostgresRepository(a.db)
	_ = notificationRepo
	// loy:endregion

	// loy:region:services
	userSvc, _ := userService.NewService(userRepo)
	_ = userSvc
	organizationSvc, _ := organizationService.NewService(organizationRepo)
	_ = organizationSvc
	billingSvc, _ := billingService.NewService(billingRepo)
	_ = billingSvc
	notificationSvc, _ := notificationService.NewService(notificationRepo)
	_ = notificationSvc
	// loy:endregion

	// loy:region:handlers
	userHandler, _ := userHttp.NewHandler(userSvc)
	organizationHandler, _ := organizationHttp.NewHandler(organizationSvc)
	billingHandler, _ := billingHttp.NewHandler(billingSvc)
	notificationHandler, _ := notificationHttp.NewHandler(notificationSvc)
	// loy:endregion

	// loy:region:routes
	if a.router != nil {
		userHandler.RegisterRoutes(a.router.Group("/api/v1"))
	}
	if a.router != nil {
		organizationHandler.RegisterRoutes(a.router.Group("/api/v1"))
	}
	if a.router != nil {
		billingHandler.RegisterRoutes(a.router.Group("/api/v1"))
	}
	if a.router != nil {
		notificationHandler.RegisterRoutes(a.router.Group("/api/v1"))
	}
	// loy:endregion

	return nil
}
