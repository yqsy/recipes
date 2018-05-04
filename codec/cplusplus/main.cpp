#include <iostream>
#include <string>
#include <assert.h>

#include <proto/query.pb.h>


int main() {
    typedef codec::Query T;

    std::string type_name = T::descriptor()->full_name();

    std::cout << type_name << std::endl;

    const ::google::protobuf::Descriptor *descriptor =
            ::google::protobuf::DescriptorPool::generated_pool()->FindMessageTypeByName(type_name);
    assert(descriptor == T::descriptor());
    std::cout << "FindMessageTypeByName() = " << descriptor << std::endl;
    std::cout << "T::descriptor()         = " << T::descriptor() << std::endl;
    std::cout << std::endl;

    const ::google::protobuf::Message *prototype =
            ::google::protobuf::MessageFactory::generated_factory()->GetPrototype(descriptor);
    assert(prototype == &T::default_instance());
    std::cout << "GetPrototype()        = " << prototype << std::endl;
    std::cout << "T::default_instance() = " << &T::default_instance() << std::endl;
    std::cout << std::endl;

    T *new_obj = dynamic_cast<T *>(prototype->New());
    assert(new_obj != NULL);
    assert(new_obj != prototype);
    assert(typeid(*new_obj) == typeid(T::default_instance()));

    std::cout << "prototype->New() = " << new_obj << std::endl;
    std::cout << std::endl;
    delete new_obj;
}
