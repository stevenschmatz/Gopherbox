//
//  GBLoginViewController.m
//  Gopherbox
//
//  Created by Steven Schmatz on 8/3/14.
//  Copyright (c) 2014 Gopherbox. All rights reserved.
//

#import "GBLoginWindowController.h"
#import "GBCursorButton.h"
#import "GBAppDelegate.h"

@interface GBLoginWindowController ()

// User interface fields
@property (weak) IBOutlet NSTextField *usernameField;
@property (weak) IBOutlet NSSecureTextField *passwordField;
@property (weak) IBOutlet GBCursorButton *loginButton;
@property (weak) IBOutlet NSProgressIndicator *loginIndicator;
@property (weak) IBOutlet GBCursorButton *signupButton;
@property (weak) IBOutlet NSTextField *signupTextField;

- (IBAction)loginInitiated:(id)sender;
- (BOOL)verifyAccountWithUsername:(NSString *)username AndPassword:(NSString *)password;

@end

@implementation GBLoginWindowController

- (id)initWithWindow:(NSWindow *)window
{
    self = [super initWithWindow:window];
    if (self) {
        self.username = @"";
        self.password = @"";
        [self.loginIndicator setHidden:YES];
    }
    return self;
}

- (void)windowDidLoad
{
    [super windowDidLoad];
    
    // Implement this method to handle any initialization after your window controller's window has been loaded from its nib file.
}

// Initiates
- (IBAction)loginInitiated:(id)sender {
    
    // Check verification
    NSString *username = self.usernameField.stringValue;
    self.username = username;
    NSString *password = self.passwordField.stringValue;
    self.password = password;
    BOOL verified = [self verifyAccountWithUsername:username AndPassword:password];
    self.verified = verified;
    
    
    if (verified) {
        [self.loginWindow close];
    }
    
    GBAppDelegate *appDelegate = [[NSApplication sharedApplication] delegate];
    [appDelegate loginVerified:verified withUsername:self.username andPassword:self.password];
    
    // At this point, the server should generate a random "session token" and store it in the database. Send to client as return value of login.
    // Login with these details
}

// verifyAccountWithUsername uses the login binary to validate the user.
- (BOOL)verifyAccountWithUsername:(NSString *)username AndPassword:(NSString *)password {
    
    [self.loginIndicator setHidden: NO];
    [self.loginIndicator startAnimation:self];
    
    NSString *binaryName = @"loginFromClient";
    NSString *launchPath = [[NSBundle mainBundle] pathForResource:binaryName ofType:nil];
    NSString *resourcesPath = [launchPath substringToIndex:launchPath.length - binaryName.length];
    
    NSArray *arguments = [NSArray arrayWithObjects:username, password, resourcesPath, nil];
    NSString *results = [self runBinaryWithName:binaryName andArguments:arguments];
    
    [self.loginIndicator stopAnimation:self];
    [self.loginIndicator setHidden: YES];
    
    
    if ([results isEqual: @"true"]) {
        return true;
    } else {
        return false;
    }
}

// runBinaryWithName runs the binary with the given name and arguments in the resources folder.
- (NSString *)runBinaryWithName:(NSString *)binaryName andArguments:(NSArray *)arguments {
    
    NSString *launchPath = [[NSBundle mainBundle] pathForResource:binaryName ofType:nil];
    
    NSTask *task;
    task = [[NSTask alloc] init];
    [task setLaunchPath: launchPath];
    
    [task setArguments: arguments];
    
    NSPipe *pipe;
    pipe = [NSPipe pipe];
    [task setStandardOutput: pipe];
    
    NSFileHandle *file;
    file = [pipe fileHandleForReading];
    
    [task launch];
    
    NSData *data;
    data = [file readDataToEndOfFile];
    
    NSString *results;
    results = [[NSString alloc] initWithData: data encoding: NSUTF8StringEncoding];
    
    return results;
}

- (IBAction)goToSignupPage:(id)sender
{
    [[NSWorkspace sharedWorkspace] openURL: [NSURL URLWithString:@"http://gopherbox.io"]];
}

@end
