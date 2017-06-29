package device

// RegistrationStream a channel that sends along new device connections after they've been upgrated to websockets
type RegistrationStream chan Connection
