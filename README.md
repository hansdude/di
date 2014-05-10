di
==

A simple golang Dependency Injection container

---

At the moment this is just a DSL parser.

---

# Example

Say your server application requires an interface for ranking products for a user, which needs an interface that knows how to talk to the database. The applicaiton also needs an interface to serialize both input and output.

Hopefully you're already thinking about applications this way. If the app needs to do something, you create an interface to talk to, implement it, and always depend on the interface rather than the implementation. Writing code this way make pieces re-usable, _unit_ testable, and easy to reason about.

But there's still a problem of wiring it all together.

Here are the parts of that example app that we need to worry about:

myapp/srv/srv.go

    package srv
    
    import (
        "myapp/encoding"
        "myapp/ranking"
    )
    
    type Server interface {
        Start()
    }
    func New(enc encoding.Encoder, ranker ranking.Ranker) Server {
        // ... implementation ...
    }
    
    
myapp/encoding/encoding.go

    package encoding
    
    type Encoder interface { /* ... */ }
    
myapp/encoding/thrift/thrift.go

    package thrift
    
    import "myapp/encoding"
    
    func New() encoding.Encoder {
        // ... implementation ...
    }
    
    
myapp/ranking/ranking.go

    package ranking
    
    type Ranker interface { /* ... */ }
    
    
myapp/ranking/slope_one/slope_one.go

    package slope_one
    
    import (
        "myapp/db"
        "myapp/ranking"
    )
    
    func New(dbAdapter db.Adapter) ranking.Ranker {
        // ... implementation ...
    }
    
    
myapp/db/db.go

    package db
    
    type Adapter interface { /* ... */ }

myapp/db/postgres/postgres.go

    package postgres
    
    import "myapp/db"
    
    func New() db.Adapter {
        // ... implementation ...
    }
    

To wire this all up, we need to do something like this:

    import (
        "myapp/srv"
        "myapp/db/postgres"
        "myapp/encoding/thrift"
        "myapp/ranking/slope_one"
    )
    
    func NewCompositionRoot() srv.Server  {
        srv.New(thrift.New(), slope_one.New(postgres.New()))
    }
    
    
    
In the future, you'll be able to wire up your application like so:

    import (
        "myapp/srv"
        "myapp/db/postgres"
        "myapp/encoding/thrift"
        "myapp/ranking/slope_one"
    )
    
    root srv.New
    slope_one.New
    reg postgres.New
    reg thrift.New
