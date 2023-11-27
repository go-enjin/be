// Copyright (C) 2023  Syrge Inc - All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Proprietary and confidential.

package permissions

func NewAccelerometer(origins ...Origin) Directive {
	return NewDirective("accelerometer", origins...)
}

func NewAmbientLightSensor(origins ...Origin) Directive {
	return NewDirective("ambient-light-sensor", origins...)
}

func NewAutoplay(origins ...Origin) Directive {
	return NewDirective("autoplay", origins...)
}

func NewBattery(origins ...Origin) Directive {
	return NewDirective("battery", origins...)
}

func NewCamera(origins ...Origin) Directive {
	return NewDirective("camera", origins...)
}

func NewDisplayCapture(origins ...Origin) Directive {
	return NewDirective("display-capture", origins...)
}

func NewDocumentDomain(origins ...Origin) Directive {
	return NewDirective("document-domain", origins...)
}

func NewEncryptedMedia(origins ...Origin) Directive {
	return NewDirective("encrypted-media", origins...)
}

func NewExecutionWhileNotRendered(origins ...Origin) Directive {
	return NewDirective("execution-while-not-rendered", origins...)
}

func NewHidden(origins ...Origin) Directive {
	return NewDirective("hidden", origins...)
}

func NewExecutionWhileOutOfViewport(origins ...Origin) Directive {
	return NewDirective("execution-while-out-of-viewport", origins...)
}

func NewFullscreen(origins ...Origin) Directive {
	return NewDirective("fullscreen", origins...)
}

func NewGamepad(origins ...Origin) Directive {
	return NewDirective("gamepad", origins...)
}

func NewGamepadconnected(origins ...Origin) Directive {
	return NewDirective("gamepadconnected", origins...)
}

func NewGeolocation(origins ...Origin) Directive {
	return NewDirective("geolocation", origins...)
}

func NewGyroscope(origins ...Origin) Directive {
	return NewDirective("gyroscope", origins...)
}

func NewHid(origins ...Origin) Directive {
	return NewDirective("hid", origins...)
}

func NewIdleDetection(origins ...Origin) Directive {
	return NewDirective("idle-detection", origins...)
}

func NewLocalFonts(origins ...Origin) Directive {
	return NewDirective("local-fonts", origins...)
}

func NewMagnetometer(origins ...Origin) Directive {
	return NewDirective("magnetometer", origins...)
}

func NewMicrophone(origins ...Origin) Directive {
	return NewDirective("microphone", origins...)
}

func NewMidi(origins ...Origin) Directive {
	return NewDirective("midi", origins...)
}

func NewPayment(origins ...Origin) Directive {
	return NewDirective("payment", origins...)
}

func NewPictureInPicture(origins ...Origin) Directive {
	return NewDirective("picture-in-picture", origins...)
}

func NewPublickeyCredentialsGet(origins ...Origin) Directive {
	return NewDirective("publickey-credentials-get", origins...)
}

func NewScreenWakeLock(origins ...Origin) Directive {
	return NewDirective("screen-wake-lock", origins...)
}

func NewSerial(origins ...Origin) Directive {
	return NewDirective("serial", origins...)
}

func NewSpeakerSelection(origins ...Origin) Directive {
	return NewDirective("speaker-selection", origins...)
}

func NewUsb(origins ...Origin) Directive {
	return NewDirective("usb", origins...)
}

func NewWebShare(origins ...Origin) Directive {
	return NewDirective("web-share", origins...)
}

func NewXrSpatialTracking(origins ...Origin) Directive {
	return NewDirective("xr-spatial-tracking", origins...)
}
