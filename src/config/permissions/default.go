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
	UserRoles: map[string]*UserRole{
		"nobody": &UserRole{
			Join:                false,
			JoinDisabledMessage: "You must be logged in to add users",

			// Ain't nobody have access
			DeviceRole: DeviceRole{
				PrivateAccessLevel: "none",
				PublicAccessLevel:  "none",
				UserAccessLevel:    "none",
				SelfAccessLevel:    "none",
			},
		},
		"user": &UserRole{
			Join:                false,
			JoinDisabledMessage: "You must be logged in as admin to add users",

			CanBePrivate: false,

			// ... Not bad
			DeviceRole: DeviceRole{
				PrivateAccessLevel: "none",
				PublicAccessLevel:  "userpublic",
				UserAccessLevel:    "userself",
				SelfAccessLevel:    "userself",
			},
		},
		"admin": &UserRole{
			Join:                true,
			JoinDisabledMessage: "Join is disabled",

			CanBePrivate: false,

			// ACCESS ALL THE THINGS
			DeviceRole: DeviceRole{
				PrivateAccessLevel: "full",
				PublicAccessLevel:  "full",
				UserAccessLevel:    "full",
				SelfAccessLevel:    "full",
			},
		},
	},

	DeviceRoles: map[string]*DeviceRole{
		"none": &DeviceRole{
			PrivateAccessLevel: "none",
			PublicAccessLevel:  "none",
			UserAccessLevel:    "none",
			SelfAccessLevel:    "full",
		},
		"reader": &DeviceRole{
			PrivateAccessLevel: "devicereader",
			PublicAccessLevel:  "devicereader",
			UserAccessLevel:    "devicereader",
			SelfAccessLevel:    "full",
		},
		"writer": &DeviceRole{
			PrivateAccessLevel: "devicewriter",
			PublicAccessLevel:  "devicewriter",
			UserAccessLevel:    "devicewriter",
			SelfAccessLevel:    "full",
		},
		"user": &DeviceRole{
			PrivateAccessLevel: "full",
			PublicAccessLevel:  "full",
			UserAccessLevel:    "full",
			SelfAccessLevel:    "full",
		},
	},

	AccessLevels: map[string]*AccessLevel{
		"userpublic": &AccessLevel{
			CanCreateUser:   false,
			CanCreateDevice: false,
			CanCreateStream: false,
			CanDeleteUser:   false,
			CanDeleteDevice: false,
			CanDeleteStream: false,

			ReadAccess:  "publicread",
			WriteAccess: "none",
		},
		"userself": &AccessLevel{
			CanCreateUser:   false,
			CanCreateDevice: true,
			CanCreateStream: true,
			CanDeleteUser:   false,
			CanDeleteDevice: true,
			CanDeleteStream: true,

			ReadAccess:  "selfread",
			WriteAccess: "selfwrite",
		},
		"devicereader": &AccessLevel{
			CanCreateUser:   false,
			CanCreateDevice: false,
			CanCreateStream: false,
			CanDeleteUser:   false,
			CanDeleteDevice: false,
			CanDeleteStream: false,

			ReadAccess:  "deviceread",
			WriteAccess: "none",
		},
		"devicewriter": &AccessLevel{
			CanCreateUser:   false,
			CanCreateDevice: false,
			CanCreateStream: false,
			CanDeleteUser:   false,
			CanDeleteDevice: false,
			CanDeleteStream: false,

			ReadAccess:  "deviceread",
			WriteAccess: "devicewrite",
		},
	},

	RWAccess: map[string]*RWAccess{
		"publicread": &RWAccess{
			CanAccessUser:                   true,
			CanAccessDevice:                 true,
			CanAccessStream:                 true,
			CanAccessNonUserEditableDevices: false,
			UserName:                        true,
			UserNickname:                    true,
			UserEmail:                       false,
			UserDescription:                 true,
			UserIcon:                        true,
			UserRoles:                       false,
			UserPublic:                      false,
			UserPassword:                    false,
			DeviceName:                      true,
			DeviceNickname:                  true,
			DeviceDescription:               true,
			DeviceIcon:                      true,
			DeviceAPIKey:                    false,
			DeviceEnabled:                   true,
			DeviceIsVisible:                 true,
			DeviceUserEditable:              false,
			DevicePublic:                    true,
			DeviceRoles:                     false,
			StreamName:                      true,
			StreamNickname:                  true,
			StreamDescription:               true,
			StreamIcon:                      true,
			StreamSchema:                    true,
			StreamEphemeral:                 true,
			StreamDownlink:                  true,
		},
		"selfwrite": &RWAccess{
			CanAccessUser:                   true,
			CanAccessDevice:                 true,
			CanAccessStream:                 true,
			CanAccessNonUserEditableDevices: false,
			UserName:                        false,
			UserNickname:                    true,
			UserEmail:                       true,
			UserDescription:                 true,
			UserIcon:                        true,
			UserRoles:                       false,
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
			DeviceRoles:                     true,
			StreamName:                      false,
			StreamNickname:                  true,
			StreamDescription:               true,
			StreamIcon:                      true,
			StreamSchema:                    true,
			StreamEphemeral:                 true,
			StreamDownlink:                  true,
		},
		"selfread": &RWAccess{
			CanAccessUser:                   true,
			CanAccessDevice:                 true,
			CanAccessStream:                 true,
			CanAccessNonUserEditableDevices: true,
			UserName:                        true,
			UserNickname:                    true,
			UserEmail:                       true,
			UserDescription:                 true,
			UserIcon:                        true,
			UserRoles:                       false,
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
			DeviceRoles:                     true,
			StreamName:                      true,
			StreamNickname:                  true,
			StreamDescription:               true,
			StreamIcon:                      true,
			StreamSchema:                    true,
			StreamEphemeral:                 true,
			StreamDownlink:                  true,
		},
		"deviceread": &RWAccess{
			CanAccessUser:                   true,
			CanAccessDevice:                 true,
			CanAccessStream:                 true,
			CanAccessNonUserEditableDevices: true,
			UserName:                        true,
			UserNickname:                    true,
			UserEmail:                       true,
			UserDescription:                 true,
			UserIcon:                        true,
			UserRoles:                       true,
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
			DeviceRoles:                     true,
			StreamName:                      true,
			StreamNickname:                  true,
			StreamDescription:               true,
			StreamIcon:                      true,
			StreamSchema:                    true,
			StreamEphemeral:                 true,
			StreamDownlink:                  true,
		},
		"devicewrite": &RWAccess{
			CanAccessUser:                   true,
			CanAccessDevice:                 true,
			CanAccessStream:                 true,
			CanAccessNonUserEditableDevices: false,
			UserName:                        false,
			UserNickname:                    true,
			UserEmail:                       true,
			UserDescription:                 true,
			UserIcon:                        true,
			UserRoles:                       false,
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
			DeviceRoles:                     false,
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
