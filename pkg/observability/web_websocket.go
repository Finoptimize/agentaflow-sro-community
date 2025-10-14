package observability

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// handleWebSocket handles WebSocket connections for real-time updates
func (wd *WebDashboard) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wd.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register connection
	wd.wsMutex.Lock()
	wd.wsConnections[conn] = true
	wd.wsMutex.Unlock()

	log.Printf("New WebSocket connection established from %s", r.RemoteAddr)

	// Send initial metrics data
	wd.sendMetricsToConnection(conn)

	// Handle incoming messages (for future interactive features)
	go wd.handleWebSocketMessages(conn)

	// Keep connection alive with ping/pong
	wd.keepConnectionAlive(conn)

	// Clean up when connection closes
	wd.wsMutex.Lock()
	delete(wd.wsConnections, conn)
	wd.wsMutex.Unlock()

	log.Printf("WebSocket connection closed from %s", r.RemoteAddr)
}

// handleWebSocketMessages processes incoming WebSocket messages
func (wd *WebDashboard) handleWebSocketMessages(conn *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WebSocket message handler panic: %v", r)
		}
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err == nil {
				wd.handleWebSocketCommand(conn, msg)
			}
		}
	}
}

// handleWebSocketCommand processes WebSocket commands
func (wd *WebDashboard) handleWebSocketCommand(conn *websocket.Conn, cmd map[string]interface{}) {
	cmdType, ok := cmd["type"].(string)
	if !ok {
		return
	}

	switch cmdType {
	case "subscribe":
		// Handle subscription to specific metrics
		wd.handleSubscription(conn, cmd)
	case "unsubscribe":
		// Handle unsubscription
		wd.handleUnsubscription(conn, cmd)
	case "get_metrics":
		// Send current metrics
		wd.sendMetricsToConnection(conn)
	case "ping":
		// Respond to ping
		wd.sendToConnection(conn, map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now(),
		})
	}
}

// handleSubscription handles metric subscription requests
func (wd *WebDashboard) handleSubscription(conn *websocket.Conn, cmd map[string]interface{}) {
	// Future enhancement: Allow clients to subscribe to specific metrics
	log.Printf("WebSocket subscription request: %v", cmd)
}

// handleUnsubscription handles metric unsubscription requests
func (wd *WebDashboard) handleUnsubscription(conn *websocket.Conn, cmd map[string]interface{}) {
	// Future enhancement: Allow clients to unsubscribe from specific metrics
	log.Printf("WebSocket unsubscription request: %v", cmd)
}

// keepConnectionAlive maintains WebSocket connection with ping/pong
func (wd *WebDashboard) keepConnectionAlive(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Printf("WebSocket ping error: %v", err)
			break
		}
	}
}

// startWebSocketBroadcast starts the background routine for broadcasting metrics
func (wd *WebDashboard) startWebSocketBroadcast() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		wd.broadcastMetrics()
	}
}

// broadcastMetrics sends current metrics to all connected WebSocket clients
func (wd *WebDashboard) broadcastMetrics() {
	wd.mu.RLock()
	metrics := DashboardMetrics{
		Timestamp:   time.Now(),
		GPUMetrics:  wd.lastMetrics,
		SystemStats: wd.calculateSystemStats(),
		CostData:    wd.lastCostData,
		Alerts:      wd.getActiveAlerts(),
		Performance: wd.calculatePerformanceMetrics(),
	}
	wd.mu.RUnlock()

	message := map[string]interface{}{
		"type": "metrics_update",
		"data": metrics,
	}

	wd.broadcastToAllConnections(message)
}

// sendMetricsToConnection sends current metrics to a specific connection
func (wd *WebDashboard) sendMetricsToConnection(conn *websocket.Conn) {
	wd.mu.RLock()
	metrics := DashboardMetrics{
		Timestamp:   time.Now(),
		GPUMetrics:  wd.lastMetrics,
		SystemStats: wd.calculateSystemStats(),
		CostData:    wd.lastCostData,
		Alerts:      wd.getActiveAlerts(),
		Performance: wd.calculatePerformanceMetrics(),
	}
	wd.mu.RUnlock()

	message := map[string]interface{}{
		"type": "metrics_update",
		"data": metrics,
	}

	wd.sendToConnection(conn, message)
}

// broadcastToAllConnections sends a message to all connected WebSocket clients
func (wd *WebDashboard) broadcastToAllConnections(message interface{}) {
	wd.wsMutex.RLock()
	connections := make([]*websocket.Conn, 0, len(wd.wsConnections))
	for conn := range wd.wsConnections {
		connections = append(connections, conn)
	}
	wd.wsMutex.RUnlock()

	for _, conn := range connections {
		wd.sendToConnection(conn, message)
	}
}

// sendToConnection sends a message to a specific WebSocket connection
func (wd *WebDashboard) sendToConnection(conn *websocket.Conn, message interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WebSocket send panic: %v", r)
			// Remove failed connection
			wd.wsMutex.Lock()
			delete(wd.wsConnections, conn)
			wd.wsMutex.Unlock()
		}
	}()

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.WriteJSON(message); err != nil {
		log.Printf("WebSocket write error: %v", err)
		// Remove failed connection
		wd.wsMutex.Lock()
		delete(wd.wsConnections, conn)
		wd.wsMutex.Unlock()
	}
}

// GetActiveConnections returns the number of active WebSocket connections
func (wd *WebDashboard) GetActiveConnections() int {
	wd.wsMutex.RLock()
	defer wd.wsMutex.RUnlock()
	return len(wd.wsConnections)
}

// BroadcastAlert sends an alert to all connected clients immediately
func (wd *WebDashboard) BroadcastAlert(alert Alert) {
	message := map[string]interface{}{
		"type": "alert",
		"data": alert,
	}

	wd.broadcastToAllConnections(message)
	log.Printf("Broadcasted alert to %d connections: %s", wd.GetActiveConnections(), alert.Message)
}

// BroadcastSystemUpdate sends a system status update to all connected clients
func (wd *WebDashboard) BroadcastSystemUpdate(update map[string]interface{}) {
	message := map[string]interface{}{
		"type": "system_update",
		"data": update,
	}

	wd.broadcastToAllConnections(message)
}

// SendNotification sends a custom notification to all connected clients
func (wd *WebDashboard) SendNotification(title, message, level string) {
	notification := map[string]interface{}{
		"type": "notification",
		"data": map[string]interface{}{
			"title":     title,
			"message":   message,
			"level":     level, // success, info, warning, error
			"timestamp": time.Now(),
		},
	}

	wd.broadcastToAllConnections(notification)
}
