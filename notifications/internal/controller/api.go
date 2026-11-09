package controller

import "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/notifications/v1"

type API struct {
	Notifications notifications.NotificationsServer
}

func New(notifications notifications.NotificationsServer) *API {
	return &API{Notifications: notifications}
}
