package permissions

// Default represents the default permissions to use in ConnectorDB. They are what is used when ConnectorDB's config has "default"
// as its "permissions" field value
var Default = Permissions{
	Version: 1,
	Watch:   true,

	// Here we disallow names that would conflict with the ConnectorDB frontend
	DisallowedNames: []string{"support", "www", "api", "app", "favicon.ico", "robots.txt", "sitemap.xml", "join", "login", "user", "admin", "nobody", "root"},

	// Allow an arbitrary number of users by default
	MaxUsers: -1,

	// Only one role is required by ConnectorDB: "nobody". This is the role automatically assigned to random
	// people that visit the web server. This is obviously not ideal.
	//
	// By Default, ConnectorDB creates 3 roles:
	// - nobody:
	//     Can't access anything - is greeted with the landing page and a login prompt
	// - user:
	//     Can create its own streams/devices, can read "public" users/devices, but cannot write
	//     anything public, and cannot read anything private.
	// - admin:
	//     Total control: Can read/write anything anywhere. Also has permissions to create new users.
	//
	// This default is very privacy-oriented, and is therefore recommended for private/family/IOT ConnectorDB instances.
	//
	// Remember that the "full" and "none" access levels are built-in (ie, don't need to be defined)
	Roles: map[string]*Role{
		"nobody": &Role{
			Join:                false,
			JoinDisabledMessage: "You must be logged in as admin to add users",

			PublicReadAccessLevel:  "none",
			PrivateReadAccessLevel: "none",
			SelfReadAccessLevel:    "none",

			PublicWriteAccessLevel:  "none",
			PrivateWriteAccessLevel: "none",
			SelfWriteAccessLevel:    "none",

			OwnerDeviceReadAccessLevel:  "none",
			OwnerDeviceWriteAccessLevel: "none",
		},
		"user": &Role{
			Join:                false,
			JoinDisabledMessage: "You must be logged in as admin to add users",

			PublicReadAccessLevel:  "publicread",
			PrivateReadAccessLevel: "none",
			SelfReadAccessLevel:    "selfread",

			PublicWriteAccessLevel:  "none",
			PrivateWriteAccessLevel: "none",
			SelfWriteAccessLevel:    "selfwrite",

			OwnerDeviceReadAccessLevel:  "selfread",
			OwnerDeviceWriteAccessLevel: "selfwrite",
		},
		"admin": &Role{
			Join:                true,
			JoinDisabledMessage: "Join is disabled",

			PublicReadAccessLevel:  "full",
			PrivateReadAccessLevel: "full",
			SelfReadAccessLevel:    "full",

			PublicWriteAccessLevel:  "full",
			PrivateWriteAccessLevel: "full",
			SelfWriteAccessLevel:    "full",

			OwnerDeviceReadAccessLevel:  "full",
			OwnerDeviceWriteAccessLevel: "full",
		},
	},

	AccessLevels: map[string]*AccessLevel{
		"publicread": &AccessLevel{
			CanAccessUser:                   true,
			CanAccessDevice:                 true,
			CanAccessStream:                 true,
			CanAccessNonUserEditableDevices: false,
			UserName:                        true,
			UserNickname:                    true,
			UserEmail:                       true,
			UserDescription:                 true,
			UserIcon:                        true,
			UserRole:                        false,
			UserPublic:                      true,
			UserPassword:                    false,
			DeviceName:                      true,
			DeviceNickname:                  true,
			DeviceDescription:               true,
			DeviceIcon:                      true,
			DeviceAPIKey:                    false,
			DeviceEnabled:                   true,
			DeviceIsVisible:                 true,
			DeviceUserEditable:              true,
			DevicePublic:                    true,
			DeviceCanReadUser:               false,
			DeviceCanReadExternal:           false,
			DeviceCanWriteUser:              false,
			DeviceCanWriteExternal:          false,
			StreamName:                      true,
			StreamNickname:                  true,
			StreamDescription:               true,
			StreamIcon:                      true,
			StreamSchema:                    true,
			StreamEphemeral:                 true,
			StreamDownlink:                  true,
		},
		"selfwrite": &AccessLevel{
			CanAccessUser:                   true,
			CanAccessDevice:                 true,
			CanAccessStream:                 true,
			CanAccessNonUserEditableDevices: false,
			UserName:                        false,
			UserNickname:                    true,
			UserEmail:                       true,
			UserDescription:                 true,
			UserIcon:                        true,
			UserRole:                        false,
			UserPublic:                      true,
			UserPassword:                    true,
			DeviceName:                      false,
			DeviceNickname:                  true,
			DeviceDescription:               true,
			DeviceIcon:                      true,
			DeviceAPIKey:                    true,
			DeviceEnabled:                   true,
			DeviceIsVisible:                 true,
			DeviceUserEditable:              false,
			DevicePublic:                    true,
			DeviceCanReadUser:               true,
			DeviceCanReadExternal:           true,
			DeviceCanWriteUser:              true,
			DeviceCanWriteExternal:          true,
			StreamName:                      false,
			StreamNickname:                  true,
			StreamDescription:               true,
			StreamIcon:                      true,
			StreamSchema:                    true,
			StreamEphemeral:                 true,
			StreamDownlink:                  true,
		},
		"selfread": &AccessLevel{
			CanAccessUser:                   true,
			CanAccessDevice:                 true,
			CanAccessStream:                 true,
			CanAccessNonUserEditableDevices: true,
			UserName:                        true,
			UserNickname:                    true,
			UserEmail:                       true,
			UserDescription:                 true,
			UserIcon:                        true,
			UserRole:                        true,
			UserPublic:                      true,
			UserPassword:                    false,
			DeviceName:                      true,
			DeviceNickname:                  true,
			DeviceDescription:               true,
			DeviceIcon:                      true,
			DeviceAPIKey:                    true,
			DeviceEnabled:                   true,
			DeviceIsVisible:                 true,
			DeviceUserEditable:              true,
			DevicePublic:                    true,
			DeviceCanReadUser:               true,
			DeviceCanReadExternal:           true,
			DeviceCanWriteUser:              true,
			DeviceCanWriteExternal:          true,
			StreamName:                      true,
			StreamNickname:                  true,
			StreamDescription:               true,
			StreamIcon:                      true,
			StreamSchema:                    true,
			StreamEphemeral:                 true,
			StreamDownlink:                  true,
		},
	},
}
